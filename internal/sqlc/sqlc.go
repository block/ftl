package sqlc

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	sqlc "github.com/sqlc-dev/sqlc/pkg/cli"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
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

func Generate(pc projectconfig.Config, mc moduleconfig.ModuleConfig) (*schema.Module, error) {
	if err := validateSQLConfigs(mc); err != nil {
		return nil, fmt.Errorf("invalid SQL config: %w", err)
	}

	cfg, err := newConfigContext(pc, mc)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLC config: %w", err)
	}

	if err := cfg.scaffoldFile(); err != nil {
		return nil, fmt.Errorf("failed to scaffold SQLC config file: %w", err)
	}

	args := []string{"generate", "--file", cfg.getSQLCConfigPath()}
	if exitCode := sqlc.Run(args); exitCode != 0 {
		return nil, fmt.Errorf("sqlc generate failed with exit code %d", exitCode)
	}

	sch, err := schema.ModuleFromProtoFile(filepath.Join(cfg.OutDir, "queries.pb"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated schema: %w", err)
	}

	return sch, nil
}

func newConfigContext(pc projectconfig.Config, mc moduleconfig.ModuleConfig) (ConfigContext, error) {
	deployDir := filepath.Clean(filepath.Join(mc.Dir, mc.DeployDir))
	schemaDir := filepath.Clean(filepath.Join(mc.Dir, mc.SQLMigrationDirectory))
	queriesDir := filepath.Clean(filepath.Join(mc.Dir, mc.SQLQueryDirectory))
	queryPaths, err := findSQLFiles(queriesDir, deployDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to find SQL files: %w", err)
	}
	schemaPaths, err := findSQLFiles(schemaDir, deployDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to find SQL files: %w", err)
	}
	plugin, err := getCachedWASMPlugin(pc)
	if err != nil {
		return ConfigContext{}, err
	}
	return ConfigContext{
		Dir:         deployDir,
		Module:      mc.Module,
		Engine:      "mysql",
		SchemaPaths: schemaPaths,
		QueryPaths:  queryPaths,
		OutDir:      deployDir,
		Plugin:      plugin,
	}, nil
}

func getCachedWASMPlugin(pc projectconfig.Config) (WASMPlugin, error) {
	pluginPath := pc.SQLCGenFTLPath()
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

func validateSQLConfigs(config moduleconfig.ModuleConfig) error {
	if config.SQLMigrationDirectory == "" {
		return fmt.Errorf("SQL schema directory is required")
	}
	if config.SQLQueryDirectory == "" {
		return fmt.Errorf("SQL query directory is required")
	}
	if _, err := os.Stat(filepath.Join(config.Dir, config.SQLMigrationDirectory)); err != nil {
		return fmt.Errorf("SQL schema directory %s does not exist: %w", config.SQLMigrationDirectory, err)
	}
	if _, err := os.Stat(filepath.Join(config.Dir, config.SQLQueryDirectory)); err != nil {
		return fmt.Errorf("SQL query directory %s does not exist: %w", config.SQLQueryDirectory, err)
	}
	if strings.HasPrefix(config.SQLMigrationDirectory, config.SQLQueryDirectory) {
		return fmt.Errorf("SQL schema directory %s cannot be a subdirectory of SQL query directory %s", config.SQLMigrationDirectory, config.SQLQueryDirectory)
	}
	if strings.HasPrefix(config.SQLQueryDirectory, config.SQLMigrationDirectory) {
		return fmt.Errorf("SQL query directory %s cannot be a subdirectory of SQL schema directory %s", config.SQLQueryDirectory, config.SQLMigrationDirectory)
	}
	return nil
}
