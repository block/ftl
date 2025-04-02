package sql

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	xslices "slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

var queryNameRegex = regexp.MustCompile(`^-- name: ([A-Za-z0-9_]+)`)

type declUniquenessData struct {
	decl   schema.Decl
	dbName string
}

type ConfigContext struct {
	Dir         string
	Module      string
	Engine      string
	SchemaPaths []string
	QueryPaths  []string
	OutDir      string
	Plugin      WASMPlugin
	Database    string
}

func (c ConfigContext) scaffoldFile() error {
	err := internal.ScaffoldZip(Files(), c.OutDir, c)
	if err != nil {
		return fmt.Errorf("failed to scaffold SQLC config file: %w", err)
	}
	return nil
}

func (c ConfigContext) getSQLCConfigPath() string {
	return filepath.Join(c.OutDir, "sqlc.yml")
}

type WASMPlugin struct {
	URL    string
	SHA256 string
}

// AddDatabaseDeclsToSchema adds Decls generated from SQL files to the schema. If the target module already exists in the schema,
// it is overwritten.
//
// Returns true if the schema was updated, false otherwise.
func AddDatabaseDeclsToSchema(ctx context.Context, projectRoot string, mc moduleconfig.AbsModuleConfig, out *schema.Schema) error {
	var cfgs []ConfigContext
	for dbName, content := range mc.SQLDatabases {
		maybeCfg, err := newConfigContext(projectRoot, mc, dbName, content)
		if err != nil {
			return fmt.Errorf("failed to extract database %s from the file system: %w", dbName, err)
		}
		cfg, ok := maybeCfg.Get()
		if !ok {
			continue
		}
		if len(cfg.SchemaPaths) > 0 && len(cfg.QueryPaths) > 0 {
			if err := cfg.scaffoldFile(); err != nil {
				return fmt.Errorf("failed to scaffold SQLC config file: %w", err)
			}
		}
		cfgs = append(cfgs, cfg)
	}

	// directories do not exist or contain no SQL files
	if len(cfgs) == 0 {
		return nil
	}

	// Generate queries for each database (one config per database)
	sch := &schema.Module{
		Name: mc.Module,
	}
	for i, m := range out.InternalModules() {
		if m.Name == mc.Module {
			out.InternalModules()[i] = sch
			break
		}
	}
 
	// sort configs so they are processed deterministically.
	// We may end up modifying the name of a generated decl to avoid conflicts when table names are the same
	// across multiple datasources. Sorting by database name ensures that the schema produced is deterministic.
	xslices.SortFunc(cfgs, func(a, b ConfigContext) int {
		return strings.Compare(a.Database, b.Database)
	})

	// tracks declarations to detect duplicates among the generated decls
	var declUniqueness = map[string]declUniquenessData{}
	for _, cfg := range cfgs {
		var err error
		if len(cfg.QueryPaths) > 0 {
			if err = exec.Command(ctx, log.Debug, ".", "ftl-sqlc", "generate", "--file", cfg.getSQLCConfigPath()).RunStderrError(ctx); err != nil {
				return fmt.Errorf("sqlc generate failed for database %s: %w", cfg.Engine, err)
			}
			sch, err = schema.ModuleFromProtoFile(filepath.Join(cfg.OutDir, "queries.pb"))
			if err != nil {
				return fmt.Errorf("failed to parse generated schema: %w", err)
			}
		}
		if err = populatePositions(sch, cfg); err != nil {
			return fmt.Errorf("failed to populate positions: %w", err)
		}
		if err = updateSchema(out, sch, cfg, declUniqueness); err != nil {
			return fmt.Errorf("failed to add queries to schema: %w", err)
		}
	}
	return nil
}

// updateSchema updates the schema with the new database decls (databases and queries).
func updateSchema(out *schema.Schema, queries *schema.Module, cfg ConfigContext, declUniqueness map[string]declUniquenessData) error {
	dbType, err := toDatabaseType(cfg.Engine)
	if err != nil {
		return err
	}
	db := &schema.Database{
		Name: cfg.Database,
		Type: dbType,
	}

	if err := updateDuplicateDeclNames(queries, cfg.Database, declUniqueness); err != nil {
		return fmt.Errorf("failed to update duplicate declaration names: %w", err)
	}
	queries.Decls = append(queries.Decls, db)

	_, err = schema.ValidateModuleInSchema(out, optional.Some(queries))
	if err != nil {
		return fmt.Errorf("failed to validate module %s: %w", queries.Name, err)
	}
	found := false
	for i, m := range out.InternalModules() {
		if m.Name == queries.Name {
			out.InternalModules()[i].Decls = append(out.InternalModules()[i].Decls, queries.Decls...)
			found = true
			break
		}
	}
	if !found {
		out.Realms[0].Modules = append(out.Realms[0].Modules, queries)
	}

	return nil
}

func toDatabaseType(engine string) (string, error) {
	switch engine {
	case moduleconfig.EnginePostgres:
		return schema.PostgresDatabaseType, nil
	case moduleconfig.EngineMySQL:
		return schema.MySQLDatabaseType, nil
	}
	return "", fmt.Errorf("invalid engine %s", engine)
}

func newConfigContext(projectRoot string, mc moduleconfig.AbsModuleConfig, dbName string, dbContent moduleconfig.DatabaseContent) (optional.Option[ConfigContext], error) {
	outDir := filepath.Join(mc.DeployDir, dbName)
	err := os.MkdirAll(outDir, 0750)
	if err != nil {
		return optional.None[ConfigContext](), fmt.Errorf("failed to create output directory %s: %w", outDir, err)
	}

	schemaDir, ok := dbContent.SchemaDir.Get()
	if !ok {
		return optional.None[ConfigContext](), nil
	}
	schemaPaths, err := findSQLFiles(filepath.Join(mc.Dir, schemaDir), outDir)
	if err != nil {
		return optional.None[ConfigContext](), fmt.Errorf("no SQL migration files found in schema directory: %w", err)
	}

	var queryPaths []string
	if queriesDir, ok := dbContent.QueriesDir.Get(); ok {
		queryPaths, err = findSQLFiles(filepath.Join(mc.Dir, queriesDir), outDir)
		if err != nil {
			return optional.None[ConfigContext](), fmt.Errorf("no SQL query files found in queries directory: %w", err)
		}
	}

	// we only need to load the plugin if there are queries
	var plugin WASMPlugin
	if len(queryPaths) > 0 {
		plugin, err = getCachedWASMPlugin(projectRoot)
		if err != nil {
			return optional.None[ConfigContext](), err
		}
	}
	return optional.Some(ConfigContext{
		Dir:         mc.DeployDir,
		Module:      mc.Module,
		Engine:      dbContent.Engine,
		SchemaPaths: schemaPaths,
		QueryPaths:  queryPaths,
		OutDir:      outDir,
		Plugin:      plugin,
		Database:    dbName,
	}), nil
}

func getCachedWASMPlugin(projectRoot string) (WASMPlugin, error) {
	pluginPath := filepath.Join(projectRoot, ".ftl", "resources", "sqlc-gen-ftl.wasm")
	if _, err := os.Stat(pluginPath); err == nil {
		return toWASMPlugin(pluginPath)
	}
	if err := extractEmbeddedFile("sqlc-gen-ftl.wasm", pluginPath); err != nil {
		return WASMPlugin{}, err
	}
	return toWASMPlugin(pluginPath)
}

func toWASMPlugin(path string) (WASMPlugin, error) {
	sha256, err := computeSHA256(path)
	if err != nil {
		return WASMPlugin{}, err
	}
	return WASMPlugin{
		URL:    fmt.Sprintf("file://%s", path),
		SHA256: sha256,
	}, nil
}

func computeSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func findSQLFiles(dir string, relativeToDir string) ([]string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []string{}, nil
	}
	relDir, err := filepath.Rel(relativeToDir, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL directory relative to %s: %w", relativeToDir, err)
	}
	var sqlFiles []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(relDir, strings.TrimPrefix(path, dir)))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk SQL files: %w", err)
	}
	return sqlFiles, nil
}

// populatePositions adds positions to sql verbs in the schema.
//
// SQLC does not provide enough information to determine the position of a verb in the source sql file.
// This is best effort.
func populatePositions(m *schema.Module, cfg ConfigContext) error {
	posMap := map[string]schema.Position{}
	for _, sqlPath := range cfg.QueryPaths {
		absPath, err := filepath.Abs(filepath.Join(cfg.OutDir, sqlPath))
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", sqlPath, err)
		}
		sql, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", absPath, err)
		}
		lines := strings.Split(string(sql), "\n")
		for i, line := range lines {
			if match := queryNameRegex.FindStringSubmatch(line); len(match) > 1 {
				posMap[strcase.ToLowerCamel(match[1])] = schema.Position{
					Filename: absPath,
					Line:     i + 1,
				}
			}
		}
	}
	dataTypes := map[string]*schema.Data{}
	for data := range slices.FilterVariants[*schema.Data](m.Decls) {
		dataTypes[data.Name] = data
	}
	for verb := range slices.FilterVariants[*schema.Verb](m.Decls) {
		pos, ok := posMap[verb.Name]
		if !ok {
			continue
		}
		verb.Pos = pos
		_ = schema.Visit(verb, func(n schema.Node, next func() error) error { //nolint:errcheck
			ref, ok := n.(*schema.Ref)
			if !ok {
				return next()
			}
			data, ok := dataTypes[ref.Name]
			if !ok {
				return next()
			}
			if data.Pos.Filename != "" && data.Pos.String() <= pos.String() {
				// multiple verbs can refer to the same data type
				// keep existing pos as it is ordered first
				return next()
			}
			data.Pos = pos
			return next()
		})
	}
	return nil
}

// updateDuplicateDeclName updates the declaration name if it already exists in another generated module
func updateDuplicateDeclNames(module *schema.Module, dbName string, declUniqueness map[string]declUniquenessData) error {
	for _, d := range module.Decls {
		if loaded, exists := declUniqueness[d.GetName()]; exists {
			if loaded.dbName == dbName {
				// if duplicated within the same database, keep the duplicate to surface a build error
				return nil
			}
			oldName := d.GetName()
			switch t := d.(type) {
			case *schema.Data:
				t.Name = cases.Title(language.English).String(dbName) + t.Name
			case *schema.Verb:
				t.Name = strings.ToLower(dbName) + cases.Title(language.English).String(t.Name)
			case *schema.Database:
				return fmt.Errorf("database %s is duplicated in module %s", t.Name, module.Name)
			default:
				return fmt.Errorf("unsupported declaration %q of type %T was generated in module %s", t.GetName(), t, module.Name)
			}
			if err := updateRefs(module, schema.RefKey{Module: module.Name, Name: oldName}, d.GetName()); err != nil {
				return err
			}
			declUniqueness[d.GetName()] = declUniquenessData{
				decl:   d,
				dbName: dbName,
			}
		} else {
			declUniqueness[d.GetName()] = declUniquenessData{
				decl:   d,
				dbName: dbName,
			}
		}
	}
	return nil
}

func updateRefs(module *schema.Module, ref schema.RefKey, newName string) error {
	err := schema.Visit(module, func(n schema.Node, next func() error) error {
		r, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}
		if r.ToRefKey() == ref {
			r.Name = newName
		}
		return next()
	})
	if err != nil {
		return fmt.Errorf("failed to update refs to duplicated generated declaration %s: %w", ref.Name, err)
	}
	return nil
}
