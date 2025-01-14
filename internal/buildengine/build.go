package buildengine

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
)

var errInvalidateDependencies = errors.New("dependencies need to be updated")

// Build a module in the given directory given the schema and module config.
//
// Plugins must use a lock file to ensure that only one build is running at a time.
//
// Returns invalidateDependenciesError if the build failed due to a change in dependencies.
func build(ctx context.Context, plugin *languageplugin.LanguagePlugin, projectConfig projectconfig.Config, bctx languageplugin.BuildContext, devMode bool, devModeEndpoints chan dev.LocalEndpoint) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx).Module(bctx.Config.Module).Scope("build")
	ctx = log.ContextWithLogger(ctx, logger)

	stubsRoot := stubsLanguageDir(projectConfig.Root(), bctx.Config.Language)
	return handleBuildResult(ctx, projectConfig, bctx.Config, result.From(plugin.Build(ctx, projectConfig.Root(), stubsRoot, bctx, devMode)), devModeEndpoints)
}

// handleBuildResult processes the result of a build
func handleBuildResult(ctx context.Context, projectConfig projectconfig.Config, c moduleconfig.ModuleConfig, eitherResult result.Result[languageplugin.BuildResult], devModeEndpoints chan dev.LocalEndpoint) (moduleSchema *schema.Module, deploy []string, err error) {
	logger := log.FromContext(ctx)
	config := c.Abs()

	result, err := eitherResult.Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build module: %w", err)
	}

	if result.InvalidateDependencies {
		return nil, nil, errInvalidateDependencies
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
		return nil, nil, errors.Join(errs...) //errtrace:skip // errtraces messes up error reporting
	}

	logger.Infof("Module built (%.2fs)", time.Since(result.StartTime).Seconds())

	migrationFiles, err := handleDatabaseMigrations(config.DeployDir, config.SQLMigrationDirectory, result.Schema)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract migrations %w", err)
	}
	result.Deploy = append(result.Deploy, migrationFiles...)
	logger.Debugf("Migrations extracted %v from %s", migrationFiles, config.SQLMigrationDirectory)

	// write schema proto to deploy directory
	schemaBytes, err := proto.Marshal(result.Schema.ToProto())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	schemaPath := projectConfig.SchemaPath(config.Module)
	err = os.MkdirAll(filepath.Dir(schemaPath), 0700)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create schema directory: %w", err)
	}
	if err := os.WriteFile(schemaPath, schemaBytes, 0600); err != nil {
		return nil, nil, fmt.Errorf("failed to write schema: %w", err)
	}
	if endpoint, ok := result.DevEndpoint.Get(); ok {
		if devModeEndpoints != nil {
			parsed, err := url.Parse(endpoint)
			if err == nil {
				devModeEndpoints <- dev.LocalEndpoint{Module: config.Module, Endpoint: *parsed, DebugPort: result.DebugPort, Language: config.Language, RunnerInfoFile: result.DevRunnerInfoFile}
			}
		}
	}
	return result.Schema, result.Deploy, nil
}
