package buildengine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/sql"
)

var errInvalidateDependencies = errors.New("dependencies need to be updated")
var errSQLError = errors.New("failed to add queries to schema")

// Build a module in the given directory given the schema and module config.
//
// Plugins must use a lock file to ensure that only one build is running at a time.
//
// Returns invalidateDependenciesError if the build failed due to a change in dependencies.
func build(ctx context.Context, projectConfig projectconfig.Config, m Module, plugin *languageplugin.LanguagePlugin, bctx languageplugin.BuildContext, devMode bool, devModeEndpoints chan dev.LocalEndpoint) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx).Module(bctx.Config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	err = sql.AddDatabaseDeclsToSchema(ctx, projectConfig.Root(), bctx.Config.Abs(), bctx.Schema)
	if err != nil {
		return nil, nil, errors.WithStack(errors.Join(errSQLError, err))
	}
	stubsRoot := stubsLanguageDir(projectConfig.Root(), bctx.Config.Language)
	return errors.WithStack3(handleBuildResult(ctx, projectConfig, m, result.From(plugin.Build(ctx, projectConfig, stubsRoot, bctx, devMode)), devMode, devModeEndpoints))
}

// handleBuildResult processes the result of a build
func handleBuildResult(ctx context.Context, projectConfig projectconfig.Config, m Module, eitherResult result.Result[languageplugin.BuildResult], devMode bool, devModeEndpoints chan dev.LocalEndpoint) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx)
	config := m.Config.Abs()

	result, err := eitherResult.Result()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build module")
	}

	if result.InvalidateDependencies {
		return nil, nil, errors.WithStack(errInvalidateDependencies)
	}

	if m.SQLError != nil {
		// Plugin has rebuilt automatically even though a SQL error occurred.
		// The build needs to fail with the current sql error
		return nil, nil, errors.WithStack(m.SQLError)
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
		return nil, nil, errors.WithStack(errors.Join(errs...))
	}

	logger.Infof("Module built (%.2fs)", time.Since(result.StartTime).Seconds())

	migrationFiles, err := handleDatabaseMigrations(ctx, config, result.Schema)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to extract migration")
	}
	result.Deploy = append(result.Deploy, migrationFiles...)
	logger.Debugf("Migrations extracted %v from %s", migrationFiles, config.SQLRootDir)

	err = handleGitCommit(ctx, config.Dir, result.Schema)
	if !devMode {
		cleanFixtures(result.Schema)
	}
	if err != nil {
		logger.Debugf("Failed to save current git commit to schema: %s", err.Error())
	}

	if endpoint, ok := result.DevEndpoint.Get(); ok {
		if devModeEndpoints != nil {
			devModeEndpoints <- dev.LocalEndpoint{Module: config.Module, Endpoint: endpoint, DebugPort: result.DebugPort, Language: config.Language, HotReloadEndpoint: result.HotReloadEndpoint.Default("")}
		}
	}
	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal schema")
	}
	schemaPath := projectConfig.SchemaPath(config.Module)
	err = os.MkdirAll(filepath.Dir(schemaPath), 0700)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create schema directory")
	}
	if err := os.WriteFile(schemaPath, schemaBytes, 0600); err != nil {
		return nil, nil, errors.Wrap(err, "failed to write schema")
	}
	return result.Schema, result.Deploy, nil
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
