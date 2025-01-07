package sqlc

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

var (
	binaryCache struct {
		path   string
		tmpDir string
		mu     sync.Mutex
	}
	pluginCache struct {
		path   string
		sha256 string
		tmpDir string
		mu     sync.Mutex
	}
	initOnce       sync.Once
	sqlcBinaryName = "sqlc"
)

// maintain a cache of the SQLC binary/WASM plugin per session of the FTL CLI
func init() {
	initOnce.Do(func() {
		cleanupFn := func() {
			binaryCache.mu.Lock()
			if binaryCache.tmpDir != "" {
				if err := os.RemoveAll(binaryCache.tmpDir); err != nil {
					fmt.Fprintf(os.Stderr, "failed to cleanup SQLC binary cache: %v\n", err)
				}
			}
			binaryCache.mu.Unlock()

			pluginCache.mu.Lock()
			if pluginCache.tmpDir != "" {
				if err := os.RemoveAll(pluginCache.tmpDir); err != nil {
					fmt.Fprintf(os.Stderr, "failed to cleanup SQLC plugin cache: %v\n", err)
				}
			}
			pluginCache.mu.Unlock()
		}
		defer cleanupFn()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanupFn()
			os.Exit(1)
		}()
	})
}

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

func (c ConfigContext) getPath() (string, error) {
	relPath, err := filepath.Rel(c.Dir, c.OutDir)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	return filepath.Join(relPath, "sqlc.yml"), nil
}

type WASMPlugin struct {
	URL    string
	SHA256 string
}

func Generate(ctx context.Context, mc moduleconfig.ModuleConfig) error {
	cfg, err := newConfigContext(ctx, mc)
	if err != nil {
		return fmt.Errorf("failed to create SQLC config: %w", err)
	}

	if err := cfg.scaffoldFile(); err != nil {
		return err
	}

	binaryPath, err := getSQLCBinary(ctx)
	if err != nil {
		return err
	}

	cfgPath, err := cfg.getPath()
	if err != nil {
		return err
	}

	cmd := exec.Command(ctx, log.Info, cfg.Dir, binaryPath, "generate", "--file", cfgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing sqlc: %w", err)
	}

	return nil
}

func newConfigContext(ctx context.Context, mc moduleconfig.ModuleConfig) (ConfigContext, error) {
	deployDir := filepath.Clean(filepath.Join(mc.Dir, mc.DeployDir))
	schemaDir := filepath.Clean(filepath.Join(mc.Dir, mc.SQLMigrationDirectory))
	relSchemaDir, err := filepath.Rel(deployDir, schemaDir)
	if err != nil {
		return ConfigContext{}, fmt.Errorf("failed to get relative schema directory: %w", err)
	}
	plugin, err := getWASMPlugin(ctx)
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

func getSQLCBinary(ctx context.Context) (string, error) {
	logger := log.FromContext(ctx)
	binaryCache.mu.Lock()
	defer binaryCache.mu.Unlock()

	if binaryCache.path != "" {
		if _, err := os.Stat(binaryCache.path); err == nil {
			return binaryCache.path, nil
		}
		logger.Warnf("cached SQLC binary no longer exists, recreating")
		binaryCache.path = ""
		if binaryCache.tmpDir != "" {
			_ = os.RemoveAll(binaryCache.tmpDir)
		}
	}

	tmpDir, err := os.MkdirTemp("", "ftl-sqlc-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	logger.Debugf("created new SQLC binary cache in %s", tmpDir)

	binaryDirName := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	binaryPath := filepath.Join(binaryDirName, sqlcBinaryName)
	extractPath := filepath.Join(tmpDir, sqlcBinaryName)
	if err := extractEmbeddedFile(binaryPath, extractPath); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to extract SQLC binary: %w", err)
	}

	if err := os.Chmod(extractPath, 0750); err != nil { //nolint:gosec
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	// verify binary is executable
	verifyCmd := exec.Command(ctx, log.Debug, filepath.Dir(extractPath), extractPath, "version")
	if err := verifyCmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("extracted SQLC binary verification failed: %w", err)
	}

	binaryCache.path = extractPath
	binaryCache.tmpDir = tmpDir
	logger.Debugf("successfully cached new SQLC binary")
	return extractPath, nil
}

func getWASMPlugin(ctx context.Context) (WASMPlugin, error) {
	logger := log.FromContext(ctx)
	pluginCache.mu.Lock()
	defer pluginCache.mu.Unlock()

	if pluginCache.path != "" {
		if _, err := os.Stat(pluginCache.path); err == nil {
			return toWASMPlugin(pluginCache.path, pluginCache.sha256), nil
		}
		pluginCache.path = ""
		pluginCache.sha256 = ""
		if pluginCache.tmpDir != "" {
			_ = os.RemoveAll(pluginCache.tmpDir)
		}
	}

	// create new plugin cache
	tmpDir, err := os.MkdirTemp("", "ftl-sqlc-plugin-*")
	if err != nil {
		return WASMPlugin{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	logger.Debugf("created new SQLC WASM plugin cache in %s", tmpDir)

	pluginPath := filepath.Join(tmpDir, "sqlc-gen-ftl.wasm")
	if err := extractEmbeddedFile("sqlc-gen-ftl.wasm", pluginPath); err != nil {
		os.RemoveAll(tmpDir)
		return WASMPlugin{}, err
	}

	pluginSHA, err := computeSHA256(pluginPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return WASMPlugin{}, err
	}

	pluginCache.path = pluginPath
	pluginCache.sha256 = pluginSHA
	pluginCache.tmpDir = tmpDir

	return toWASMPlugin(pluginPath, pluginSHA), nil
}

func toWASMPlugin(path, sha string) WASMPlugin {
	return WASMPlugin{
		URL:    fmt.Sprintf("file://%s", path),
		SHA256: sha,
	}
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
