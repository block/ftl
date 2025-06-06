package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/profiles"
	"github.com/block/ftl/internal/watch"
)

type newSQLCmd struct {
	engine     string // Set by parent command
	Datasource string `arg:"" help:"The qualified name of the datasource in the form module.datasource to create. If the module is not specified FTL will attempt to infer it from the current working directory."`

	DevDirs []string `help:"Module directories that FTL Dev is discovering modules in" env:"FTL_DEV_DIRS" hidden:""`
}

func (i newSQLCmd) Run(ctx context.Context, projectConfig profiles.ProjectConfig) error {
	var searchDirs []string
	if len(i.DevDirs) > 0 {
		searchDirs = i.DevDirs
	} else {
		searchDirs = projectConfig.AbsModuleDirs()
	}
	modules, err := watch.DiscoverModules(ctx, searchDirs)
	if err != nil {
		return errors.Wrap(err, "could not discover modules")
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
		return errors.Errorf("invalid datasource %q, must be in the form module.datasource", i.Datasource)
	}
	if module == nil {
		return errors.Errorf("could not find module %q", parts[0])
	}

	// Validate engine type
	_, err = moduleconfig.ToEngineType(i.engine)
	if err != nil {
		return errors.Wrapf(err, "invalid engine type %q", i.engine)
	}

	language := module.Language
	defaults, err := languageplugin.GetModuleConfigDefaults(ctx, language, module.Dir)
	if err != nil {
		return errors.Wrapf(err, "could not get module config defaults for language %q", language)
	}

	sqlRootDir := defaults.SQLRootDir
	if sqlRootDir == "" {
		return errors.Errorf("no SQL root directory configured for language %q", language)
	}

	dbDir := filepath.Join(module.Dir, sqlRootDir, i.engine, dsName)
	schemaDir := filepath.Join(dbDir, "schema")
	queriesDir := filepath.Join(dbDir, "queries")

	if stat, err := os.Stat(schemaDir); err == nil && stat.IsDir() {
		entries, err := os.ReadDir(schemaDir)
		if err != nil {
			return errors.Wrapf(err, "could not read schema directory at %s", schemaDir)
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
		return errors.Wrapf(err, "could not create schema directory at %s", schemaDir)
	}
	if err := os.MkdirAll(queriesDir, 0750); err != nil {
		return errors.Wrapf(err, "could not create queries directory at %s", queriesDir)
	}

	// Create initial migration file
	timestamp := time.Now().UTC().Format("20060102150405")
	migrationName := fmt.Sprintf("%s_init.sql", timestamp)
	migrationPath := filepath.Join(schemaDir, migrationName)
	migrationFile, err := os.Create(migrationPath)
	if err != nil {
		return errors.Wrapf(err, "could not create migration file at %s", migrationPath)
	}
	defer migrationFile.Close()

	if _, err := migrationFile.WriteString(migrationTemplate); err != nil {
		return errors.Wrapf(err, "could not write to migration file at %s", migrationPath)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Created SQL database structure for %s in module %q", dsName, module.Module)
	fmt.Printf("Created SQL database structure at %s\n", dbDir)
	fmt.Printf("Initial migration file: %s\n", migrationPath)
	return nil
}
