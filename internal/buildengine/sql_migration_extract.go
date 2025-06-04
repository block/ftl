package buildengine

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/moduleconfig"
)

// ExtractSQLMigrations extracts all migrations from the given directory and returns the updated schema and a list of migration files to deploy.
func extractSQLMigrations(ctx context.Context, cfg moduleconfig.AbsModuleConfig, sch *schema.Module, targetDir string) ([]string, error) {
	logger := log.FromContext(ctx)
	ret := []string{}
	for db := range slices.FilterVariants[*schema.Database](sch.Decls) {
		logger.Debugf("Processing migrations for %s", db.Name)
		dbContent, ok := cfg.SQLDatabases[db.Name]
		if !ok {
			logger.Debugf("No DB content for %s", db.Name)
			continue
		}
		schemaDir, ok := dbContent.SchemaDir.Get()
		if !ok {
			logger.Debugf("No schema content for %s", db.Name)
			continue
		}
		fileName := db.Name + ".tar"
		target := filepath.Join(targetDir, fileName)
		schemaDir = filepath.Join(cfg.Dir, schemaDir)
		logger.Debugf("Reading migrations from %s", schemaDir)
		err := createMigrationTarball(schemaDir, target)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create migration tar %s", schemaDir)
		}
		digest, err := sha256.SumFile(target)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read migration tar for sha256 %s", schemaDir)
		}
		db.Metadata = append(db.Metadata, &schema.MetadataSQLMigration{Digest: digest.String()})
		ret = append(ret, fileName)
	}
	return ret, nil
}

func createMigrationTarball(migrationDir string, target string) error {
	// Create the tar file
	tarFile, err := os.Create(target)
	if err != nil {
		return errors.Wrap(err, "failed to create tar file")
	}
	defer tarFile.Close()

	// Create a new tar writer
	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	// Read the directory
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	// Sort files alphabetically
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Set the Unix epoch time
	epoch := time.Unix(0, 0)

	// Add files to the tarball
	for _, file := range files {

		filePath := filepath.Join(migrationDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			return errors.Wrap(err, "failed to stat file")
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return errors.Wrap(err, "failed to create tar header")
		}
		header.Name = file.Name()
		header.ModTime = epoch
		header.AccessTime = epoch
		header.ChangeTime = epoch

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "failed to write header")
		}

		// Write file content
		if !info.IsDir() {
			fileContent, err := os.Open(filePath)
			if err != nil {
				return errors.Wrap(err, "failed to open file")
			}
			defer fileContent.Close()

			if _, err := io.Copy(tw, fileContent); err != nil {
				return errors.Wrap(err, "failed to write file content")
			}
		}
	}
	return nil
}

func handleDatabaseMigrations(ctx context.Context, cfg moduleconfig.AbsModuleConfig, module *schema.Module) ([]string, error) {
	target := filepath.Join(cfg.DeployDir, "migrations")
	err := os.MkdirAll(target, 0770) // #nosec
	if err != nil {
		return nil, errors.Wrap(err, "failed to create migration directory")
	}
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting SQL migrations into %s", target)
	migrations, err := extractSQLMigrations(ctx, cfg, module, target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract migrations")
	}
	relativeFiles := []string{}
	for _, file := range migrations {
		filePath := filepath.Join("migrations", file)
		relativeFiles = append(relativeFiles, filePath)
	}
	return relativeFiles, nil
}
