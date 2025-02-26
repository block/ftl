package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/watch"
)

type newSQLCmd struct {
	engine     string // Set by parent command
	Datasource string `arg:"" help:"The qualified name of the datasource in the form module.datasource to create. If the module is not specified FTL will attempt to infer it from the current working directory."`
}

func newNewSQLCmd(engine string) newSQLCmd {
	return newSQLCmd{engine: engine}
}

func (i newSQLCmd) Run(ctx context.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	modules, err := watch.DiscoverModules(ctx, []string{dir})
	if err != nil {
		return fmt.Errorf("could not discover modules: %w", err)
	}
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

	// Validate engine type
	_, err = moduleconfig.ToEngineType(i.engine)
	if err != nil {
		return fmt.Errorf("invalid engine type %q: %w", i.engine, err)
	}

	language := module.Language
	plugin, err := languageplugin.CreateLanguagePlugin(ctx, language)
	if err != nil {
		return fmt.Errorf("could not create plugin for language %q: %w", language, err)
	}
	defaults, err := plugin.ModuleConfigDefaults(ctx, module.Dir)
	if err != nil {
		return fmt.Errorf("could not get module config defaults for language %q: %w", language, err)
	}

	sqlRootDir := defaults.SQLRootDir
	if sqlRootDir == "" {
		return fmt.Errorf("no SQL root directory configured for language %q", language)
	}

	dbDir := filepath.Join(module.Dir, sqlRootDir, i.engine, dsName)
	schemaDir := filepath.Join(dbDir, "schema")
	queriesDir := filepath.Join(dbDir, "queries")

	if stat, err := os.Stat(schemaDir); err == nil && stat.IsDir() {
		entries, err := os.ReadDir(schemaDir)
		if err != nil {
			return fmt.Errorf("could not read schema directory at %s: %w", schemaDir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				fmt.Printf("Database %q already exists at %s\n", dsName, dbDir)
				return nil
			}
		}
	}

	// Create directories
	if err := os.MkdirAll(schemaDir, 0750); err != nil {
		return fmt.Errorf("could not create schema directory at %s: %w", schemaDir, err)
	}
	if err := os.MkdirAll(queriesDir, 0750); err != nil {
		return fmt.Errorf("could not create queries directory at %s: %w", queriesDir, err)
	}

	// Create initial migration file
	timestamp := time.Now().UTC().Format("20060102150405")
	migrationName := fmt.Sprintf("%s_init.sql", timestamp)
	migrationPath := filepath.Join(schemaDir, migrationName)
	migrationFile, err := os.Create(migrationPath)
	if err != nil {
		return fmt.Errorf("could not create migration file at %s: %w", migrationPath, err)
	}
	defer migrationFile.Close()

	// Write initial migration template
	if _, err := migrationFile.WriteString("-- migrate:up\n-- Add your initial schema here\n\n-- migrate:down\n-- Add rollback SQL here\n"); err != nil {
		return fmt.Errorf("could not write to migration file at %s: %w", migrationPath, err)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Created SQL database structure for %s in module %q", dsName, module.Module)
	fmt.Printf("Created SQL database structure at %s\n", dbDir)
	fmt.Printf("Initial migration file: %s\n", migrationPath)
	return nil
}
