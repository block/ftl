package buildengine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/sql"
	"github.com/block/ftl/internal/watch"
)

const (
	FTLFullSchemaPath = "ftl-full-schema.pb"
)

var errInvalidateDependencies = errors.New("dependencies need to be updated")
var errSQLError = errors.New("failed to add queries to schema")

// Build a module in the given directory given the schema and module config.
//
// Plugins must use a lock file to ensure that only one build is running at a time.
//
// Returns invalidateDependenciesError if the build failed due to a change in dependencies.
func build(ctx context.Context, projectConfig projectconfig.Config, m Module, plugin *languageplugin.LanguagePlugin, fileTransaction watch.ModifyFilesTransaction, bctx languageplugin.BuildContext, devMode bool, devModeEndpoints chan dev.LocalEndpoint) (moduleSchema *schema.Module, tmpDeployDir string, deployPaths []string, err error) {
	logger := log.FromContext(ctx).Module(bctx.Config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	err = sql.AddDatabaseDeclsToSchema(ctx, projectConfig.Root(), bctx.Config.Abs(), bctx.Schema)
	if err != nil {
		return nil, "", nil, errors.WithStack(errors.Join(errSQLError, err))
	}

	stubsRoot := stubsLanguageDir(projectConfig.Root(), bctx.Config.Language)
	moduleSchema, tmpDeployDir, deployPaths, err = handleBuildResult(ctx, projectConfig, m, fileTransaction, result.From(plugin.Build(ctx, projectConfig, stubsRoot, bctx, devMode)), devMode, devModeEndpoints, optional.Some(bctx.Schema))
	if err != nil {
		return nil, "", nil, errors.WithStack(err)
	}
	return moduleSchema, tmpDeployDir, deployPaths, nil
}

// handleBuildResult processes the result of a build
func handleBuildResult(ctx context.Context, projectConfig projectconfig.Config, m Module, fileTransaction watch.ModifyFilesTransaction, eitherResult result.Result[languageplugin.BuildResult], devMode bool, devModeEndpoints chan dev.LocalEndpoint, schemaOpt optional.Option[*schema.Schema]) (moduleSchema *schema.Module, tmpDeployDir string, deployPaths []string, err error) {
	logger := log.FromContext(ctx)
	config := m.Config.Abs()

	result, err := eitherResult.Result()
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to build module")
	}

	if len(result.ModifiedFiles) > 0 {
		logger.Infof("Modified files: %v", result.ModifiedFiles)
	}
	if err := fileTransaction.ModifiedFiles(result.ModifiedFiles...); err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to apply modified files")
	}

	if result.InvalidateDependencies {
		return nil, "", nil, errors.WithStack(errInvalidateDependencies)
	}

	if m.SQLError != nil {
		// Plugin has rebuilt automatically even though a SQL error occurred.
		// The build needs to fail with the current sql error
		return nil, "", nil, errors.WithStack(m.SQLError)
	}

	var errs []error
	for _, e := range result.Errors {
		if e.Level == builderrors.WARN {
			logger.Log(log.Entry{Level: log.Warn, Message: e.Error(), Error: e})
			continue
		}
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return nil, "", nil, errors.WithStack(errors.Join(errs...))
	}

	logger.Infof("Module built (%.2fs)", time.Since(result.StartTime).Seconds())

	migrationFiles, err := handleDatabaseMigrations(ctx, config, result.Schema)
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to extract migration")
	}
	result.Deploy = append(result.Deploy, migrationFiles...)
	logger.Debugf("Migrations extracted %v from %s", migrationFiles, config.SQLRootDir)

	if schema, ok := schemaOpt.Get(); ok {
		if realm, ok := schema.FirstInternalRealm().Get(); ok {
			realm.UpsertModule(result.Schema)
		}
		data, err := proto.Marshal(schema.ToProto())
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "failed to marshal schema files")
		}
		path := filepath.Join(config.DeployDir, FTLFullSchemaPath)
		err = os.WriteFile(path, data, 0644) //nolint
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "failed to write full schema")
		}
		result.Deploy = append(result.Deploy, FTLFullSchemaPath)
	}

	tmpDeployDir, deployPaths, err = copyArtifacts(ctx, config.Module, config.DeployDir, result.Deploy)
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to copy deploy artifacts")
	}

	err = handleGitCommit(ctx, config.Dir, result.Schema)
	if !devMode {
		cleanFixtures(result.Schema)
	}
	if err != nil {
		logger.Debugf("Failed to save current git commit to schema: %s", err.Error())
	}

	if endpoint, ok := result.DevEndpoint.Get(); ok {
		var version int64
		if v, ok := result.HotReloadVersion.Get(); ok {
			version = v
		}
		if devModeEndpoints != nil {
			devModeEndpoints <- dev.LocalEndpoint{Module: config.Module, Endpoint: endpoint, DebugPort: result.DebugPort, Language: config.Language, HotReloadEndpoint: result.HotReloadEndpoint.Default(""), Version: version}
		}
	}
	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to marshal schema")
	}
	schemaPath := projectConfig.SchemaPath(config.Module)
	err = os.MkdirAll(filepath.Dir(schemaPath), 0700)
	if err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to create schema directory")
	}
	if err := os.WriteFile(schemaPath, schemaBytes, 0600); err != nil {
		return nil, "", nil, errors.Wrap(err, "failed to write schema")
	}
	return result.Schema, tmpDeployDir, deployPaths, nil
}

// cleanFixtures removes fixtures from the schema, it is invoked for non-dev mode builds
func cleanFixtures(module *schema.Module) {
	module.Decls = slices.Filter(module.Decls, func(d schema.Decl) bool {
		if verb, ok := d.(*schema.Verb); ok {
			for _, md := range verb.Metadata {
				if _, ok := md.(*schema.MetadataFixture); ok {
					return false
				}
			}
		}
		return true
	})
}

func handleGitCommit(ctx context.Context, dir string, module *schema.Module) error {
	commit, err := exec.Capture(ctx, dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return errors.Wrap(err, "failed to get git commit")
	}
	status, err := exec.Capture(ctx, dir, "git", "status", "--porcelain")
	if err != nil {
		return errors.Wrap(err, "failed to get git status")
	}
	origin, err := exec.Capture(ctx, dir, "git", "remote", "get-url", "origin")
	if err != nil {
		return errors.Wrap(err, "failed to get git origin url")
	}
	module.Metadata = append(module.Metadata, &schema.MetadataGit{Repository: strings.TrimSpace(string(origin)), Commit: strings.TrimSpace(string(commit)), Dirty: strings.TrimSpace(string(status)) != ""})
	return nil
}

func copyArtifacts(ctx context.Context, moduleName, root string, pathPatterns []string) (tmpDeployDir string, deployFiles []string, err error) {
	tmpDir, err := os.MkdirTemp("", "ftl-"+moduleName+"-artifacts")
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to create temporary directory")
	}

	paths, err := findFilesToDeploy(root, pathPatterns)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to find files to deploy")
	}
	absPaths := make([]string, 0, len(paths))
	for _, srcPath := range paths {
		relSrcPath, err := filepath.Rel(root, srcPath)
		if err != nil {
			return "", nil, errors.Wrapf(err, "failed to get relative path for deploy artifact at %s", srcPath)
		}
		dstPath := filepath.Join(tmpDir, relSrcPath)

		err = os.MkdirAll(filepath.Dir(dstPath), 0700)
		if err != nil {
			return "", nil, errors.Wrapf(err, "failed to create directory for %s", dstPath)
		}
		// Use cp -p to preserve permissions
		if err = exec.Command(ctx, log.Debug, ".", "cp", "-p", srcPath, dstPath).RunStderrError(ctx); err != nil {
			return "", nil, errors.Wrapf(err, "failed to copy file from %s to %s", srcPath, dstPath)
		}
		absPaths = append(absPaths, dstPath)
	}
	return tmpDir, absPaths, nil
}

// findFilesToDeploy returns a list of files to deploy for the given module.
func findFilesToDeploy(root string, deploy []string) ([]string, error) {
	var out []string
	for _, f := range deploy {
		file := filepath.Clean(filepath.Join(root, f))
		if !strings.HasPrefix(file, root) {
			return nil, errors.Errorf("deploy path %q is not beneath deploy directory %q", file, root)
		}
		info, err := os.Stat(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if info.IsDir() {
			dirFiles, err := findFilesInDir(file)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			out = append(out, dirFiles...)
		} else {
			out = append(out, file)
		}
	}
	return out, nil
}

func findFilesInDir(dir string) ([]string, error) {
	var out []string
	return out, errors.WithStack(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if info.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	}))
}
