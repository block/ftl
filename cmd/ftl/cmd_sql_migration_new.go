package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/amacneil/dbmate/v2/pkg/dbutil"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/buildengine/languageplugin"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/watch"
)

type migrationSQLCmd struct {
	Datasource string `arg:"" help:"The qualified name of the datasource in the form module.datasource to create the migration for. If the module is not specified FTL will attempt to infer it from the current working directory." predictor:"databases"`
	Name       string `arg:"" help:"Name of the migration, this will be included in the migration file name."`

	DevDirs []string `help:"Module directories that FTL Dev is discovering modules in" env:"FTL_DEV_DIRS" hidden:""`
}

func (i migrationSQLCmd) Run(ctx context.Context, projectConfig projectconfig.Config) error {
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
	var migrationDir string
	var found bool
	if sqlDirs, ok := module.SQLDatabases[dsName]; ok {
		migrationDir, found = sqlDirs.SchemaDir.Get()
	}
	if migrationDir == "" || !found {
		language := module.Language
		defaults, err := languageplugin.GetModuleConfigDefaults(ctx, language, module.Dir)
		if err != nil {
			return errors.Wrapf(err, "could not get module config defaults for language %q", language)
		}
		valid, databases, err := moduleconfig.ValidateSQLRoot(module.Dir, defaults.SQLRootDir)
		if err != nil {
			return errors.Wrapf(err, "could not locate SQL migration directory for %q in %q", dsName, defaults.SQLRootDir)
		}
		if !valid {
			return errors.Errorf("invalid SQL root directory %q", defaults.SQLRootDir)
		}
		if sqlDirs, ok := databases[dsName]; ok {
			migrationDir, found = sqlDirs.SchemaDir.Get()
		}
		if migrationDir == "" || !found {
			return errors.Errorf("could not get SQL migration directory for datasource %q", dsName)
		}
	}
	migrationDir = filepath.Join(module.Dir, migrationDir)

	logger := log.FromContext(ctx)
	logger.Debugf("Creating DBMate SQL migration %s in module %q in %s", i.Name, module.Module, migrationDir)
	migrationPath, err := newMigration(migrationDir, i.Name)
	if err != nil {
		return errors.Wrap(err, "failed to create migration")
	}
	fmt.Printf("Created migration at %s\n", migrationPath)
	return nil
}

// Purposely use a template string which is invalid so that autorebuilds do not deploy the empty migration
// before the user is able to enter their own migration
const migrationTemplate = "-- migrate:up\n\n\n-- migrate:down\n\n"

// newMigration creates a new migration file and returns the path
func newMigration(dir, name string) (string, error) {
	// new migration name
	timestamp := time.Now().UTC().Format("20060102150405")
	if name == "" {
		return "", errors.Errorf("migration name required")
	}
	name = fmt.Sprintf("%s_%s.sql", timestamp, name)

	// check file does not already exist
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return "", errors.Errorf("migration file already exists: %s", path)
	}

	// write new migration
	file, err := os.Create(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not create migration file at %s", path)
	}

	defer dbutil.MustClose(file)
	if _, err := file.WriteString(migrationTemplate); err != nil {
		return "", errors.Wrapf(err, "could not write to migration file at %s", path)
	}
	return path, nil
}
