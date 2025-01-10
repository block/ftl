package sqlc

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	sqlc "github.com/sqlc-dev/sqlc/pkg/cli"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

type ConfigContext struct {
	Dir        string
	Module     string
	Engine     string
	SchemaDir  string
	QueriesDir string
	OutDir     string
	Plugin     WASMPlugin
}

func (c ConfigContext) scaffoldFile() error {
	err := internal.ScaffoldZip(sqlcTemplateFiles(), c.OutDir, c)
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
	relSchemaDir, err := filepath.Rel(deployDir, schemaDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to get relative schema directory: %w", err)
	}

	plugin, err := getCachedWASMPlugin(pc)
	if err != nil {
		return ConfigContext{}, err
	}
	return ConfigContext{
		Dir:        deployDir,
		Module:     mc.Module,
		Engine:     "mysql",
		SchemaDir:  relSchemaDir,
		QueriesDir: filepath.Join(relSchemaDir, "queries"),
		OutDir:     deployDir,
		Plugin:     plugin,
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

func sqlcTemplateFiles() *zip.Reader {
	return internal.ZipRelativeToCaller("template")
}
