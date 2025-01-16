package sqlc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

type ConfigContext struct {
	Dir         string
	Module      string
	Engine      string
	SchemaPaths string
	QueryPaths  string
	OutDir      string
	Plugin      WASMPlugin
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

// AddQueriesToSchema adds Decls generated from SQL files to the schema. If the target module already exists in the schema,
// it is overwritten.
//
// Returns true if the schema was updated, false otherwise.
func AddQueriesToSchema(ctx context.Context, projectRoot string, mc moduleconfig.AbsModuleConfig, out *schema.Schema) (bool, error) {
	if !hasQueries(mc) {
		return false, nil
	}

	if err := validateSQLConfigs(mc); err != nil {
		return false, fmt.Errorf("invalid SQL config: %w", err)
	}

	cfg, err := newConfigContext(projectRoot, mc)
	if err != nil {
		return false, fmt.Errorf("failed to create SQLC config: %w", err)
	}
	// directories exist but contain no SQL files
	if cfg.QueryPaths == "" || cfg.SchemaPaths == "" {
		return false, nil
	}

	if err := cfg.scaffoldFile(); err != nil {
		return false, fmt.Errorf("failed to scaffold SQLC config file: %w", err)
	}

	if err := exec.Command(ctx, log.Debug, ".", "ftl-sqlc", "generate", "--file", cfg.getSQLCConfigPath()).RunBuffered(ctx); err != nil {
		return false, fmt.Errorf("sqlc generate failed: %w", err)
	}

	sch, err := schema.ModuleFromProtoFile(filepath.Join(cfg.OutDir, "queries.pb"))
	if err != nil {
		return false, fmt.Errorf("failed to parse generated schema: %w", err)
	}

	found := false
	for i, m := range out.Modules {
		if m.Name == sch.Name {
			out.Modules[i] = sch
			found = true
			break
		}
	}
	if !found {
		out.Modules = append(out.Modules, sch)
	}

	return true, nil
}

func hasQueries(config moduleconfig.AbsModuleConfig) bool {
	if config.SQLMigrationDirectory == "" || config.SQLQueryDirectory == "" {
		return false
	}
	if _, err := os.Stat(config.SQLMigrationDirectory); err != nil {
		return false
	}
	if _, err := os.Stat(config.SQLQueryDirectory); err != nil {
		return false
	}
	return true
}

func newConfigContext(projectRoot string, mc moduleconfig.AbsModuleConfig) (ConfigContext, error) {
	queryPaths, err := findSQLFiles(mc.SQLQueryDirectory, mc.DeployDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to find SQL files: %w", err)
	}
	schemaPaths, err := findSQLFiles(mc.SQLMigrationDirectory, mc.DeployDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to find SQL files: %w", err)
	}
	plugin, err := getCachedWASMPlugin(projectRoot)
	if err != nil {
		return ConfigContext{}, err
	}
	return ConfigContext{
		Dir:         mc.DeployDir,
		Module:      mc.Module,
		Engine:      "mysql",
		SchemaPaths: schemaPaths,
		QueryPaths:  queryPaths,
		OutDir:      mc.DeployDir,
		Plugin:      plugin,
	}, nil
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

func findSQLFiles(dir string, relativeToDir string) (string, error) {
	relDir, err := filepath.Rel(relativeToDir, dir)
	if err != nil {
		return "", fmt.Errorf("failed to get SQL directory relative to %s: %w", relativeToDir, err)
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
		return "", fmt.Errorf("failed to walk SQL files: %w", err)
	}
	return strings.Join(sqlFiles, ","), nil
}

func validateSQLConfigs(config moduleconfig.AbsModuleConfig) error {
	if config.SQLMigrationDirectory == "" {
		return fmt.Errorf("SQL schema directory is required")
	}
	if config.SQLQueryDirectory == "" {
		return fmt.Errorf("SQL query directory is required")
	}
	if _, err := os.Stat(config.SQLMigrationDirectory); err != nil {
		return fmt.Errorf("SQL schema directory %s does not exist: %w", config.SQLMigrationDirectory, err)
	}
	if _, err := os.Stat(config.SQLQueryDirectory); err != nil {
		return fmt.Errorf("SQL query directory %s does not exist: %w", config.SQLQueryDirectory, err)
	}
	if isSubPath(config.SQLMigrationDirectory, config.SQLQueryDirectory) {
		return fmt.Errorf("SQL schema directory %s cannot be a subdirectory of SQL query directory %s", config.SQLMigrationDirectory, config.SQLQueryDirectory)
	}
	if isSubPath(config.SQLQueryDirectory, config.SQLMigrationDirectory) {
		return fmt.Errorf("SQL query directory %s cannot be a subdirectory of SQL schema directory %s", config.SQLQueryDirectory, config.SQLMigrationDirectory)
	}
	return nil
}

func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != "."
}
