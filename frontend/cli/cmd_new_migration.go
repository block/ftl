package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"

	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/watch"
)

type newSQLMigrationCmd struct {
	Datasource string `arg:"" help:"The qualified name of the datasource in the form module.datasource to create the migration for. If the module is not specified FTL will attempt to infer it from the current working directory."`
	Name       string `arg:"" help:"Name of the migration, this will be included in the migration file name."`
}

func (i newSQLMigrationCmd) Run(ctx context.Context) error {

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	modules, err := watch.DiscoverModules(ctx, []string{dir})
	var module *moduleconfig.UnvalidatedModuleConfig
	parts := strings.Split(i.Datasource, ".")
	var dsName string
	if len(parts) == 1 && len(modules) == 1 {
		module = &modules[0]
		dsName = parts[0]
	} else if len(parts) == 2 {
		for i := range modules {
			if modules[i].Module == parts[0] {
				module = &modules[i]
				break
			}
		}
		dsName = parts[1]
	} else {
		return fmt.Errorf("invalid datasource %q, must be in the form module.datasource", i.Datasource)
	}
	if module == nil {
		return fmt.Errorf("could not find module %q", parts[0])
	}

	if err != nil {
		return fmt.Errorf("could not discover modules: %w", err)
	}
	migrationDir := module.SQLMigrationDirectory
	if migrationDir == "" {
		language := module.Language
		plugin, err := languageplugin.CreateLanguagePlugin(ctx, language)
		if err != nil {
			return fmt.Errorf("could not create plugin for language %q: %w", language, err)
		}
		defaults, err := plugin.ModuleConfigDefaults(ctx, module.Dir)
		if err != nil {
			return fmt.Errorf("could not get module config defaults for language %q: %w", language, err)
		}
		migrationDir = defaults.SQLMigrationDir
	}
	migrationDir = filepath.Join(module.Dir, migrationDir, dsName)

	logger := log.FromContext(ctx)
	logger.Debugf("Creating DBMate SQL migration %s in module %q in %s", i.Name, module.Module, migrationDir)
	db := dbmate.New(&url.URL{})
	db.MigrationsDir = []string{migrationDir}
	err = db.NewMigration(i.Name)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}
	return nil
}
