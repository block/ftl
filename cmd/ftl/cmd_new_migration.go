package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbutil"

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
	migrationPath, err := newMigration(migrationDir, i.Name)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}
	fmt.Printf("Created migration at %s\n", migrationPath)
	return nil
}

// Purposely use a template string which is invalid so that autorebuilds do not deploy the empty migration
// before the user is able to enter their own migration
const migrationTemplate = "-- migrate:up\nPut migration here\n\n-- migrate:down\n\n"

// newMigration creates a new migration file and returns the path
func newMigration(dir, name string) (string, error) {
	// new migration name
	timestamp := time.Now().UTC().Format("20060102150405")
	if name == "" {
		return "", fmt.Errorf("migration name required")
	}
	name = fmt.Sprintf("%s_%s.sql", timestamp, name)

	// create migrations dir if missing
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("could not create migrations directory at %s: %w", dir, err)
	}

	// check file does not already exist
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("migration file already exists: %s", path)
	}

	// write new migration
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("could not create migration file at %s: %w", path, err)
	}

	defer dbutil.MustClose(file)
	if _, err := file.WriteString(migrationTemplate); err != nil {
		return "", fmt.Errorf("could not write to migration file at %s: %w", path, err)
	}
	return path, nil
}
