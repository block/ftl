package buildengine

import (
	"github.com/block/ftl/internal/log"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/scaffolder"
)

func TestExtractMigrations(t *testing.T) {
	t.Run("Valid migrations", func(t *testing.T) {
		// Setup
		tmpDir := t.TempDir()
		err := scaffolder.Scaffold(filepath.Join("testdata", "database"), tmpDir, nil)
		assert.NoError(t, err)

		targetDir := t.TempDir()

		// Define schema with a database declaration
		db := &schema.Database{Name: "testdb"}
		sch := &schema.Module{Decls: []schema.Decl{db}}

		// Test
		files, err := extractSQLMigrations(log.ContextWithNewDefaultLogger(t.Context()), getAbsModuleConfig(t, tmpDir, "db"), sch, targetDir)
		assert.NoError(t, err)

		// Validate results
		targetFile := filepath.Join(targetDir, "testdb.tar")
		assert.Equal(t, targetFile, filepath.Join(targetDir, files[0]))

		// Validate the database metadata
		assert.Equal(t, 1, len(db.Metadata))
		migrationMetadata, ok := db.Metadata[0].(*schema.MetadataSQLMigration)
		assert.True(t, ok)
		expectedDigest, err := sha256.SumFile(targetFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedDigest.String(), migrationMetadata.Digest)
	})

	t.Run("Empty migrations directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		sch := &schema.Module{Decls: []schema.Decl{}}

		files, err := extractSQLMigrations(log.ContextWithNewDefaultLogger(t.Context()), getAbsModuleConfig(t, tmpDir, "db"), sch, t.TempDir())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(files))
	})

	t.Run("Missing migrations directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		sch := &schema.Module{Decls: []schema.Decl{}}

		files, err := extractSQLMigrations(log.ContextWithNewDefaultLogger(t.Context()), getAbsModuleConfig(t, tmpDir, "/non/existent/dir"), sch, t.TempDir())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(files))
	})

}

func getAbsModuleConfig(t *testing.T, moduleDir string, sqlRootDir string) moduleconfig.AbsModuleConfig {
	mc, err := moduleconfig.UnvalidatedModuleConfig{
		Dir:        moduleDir,
		DeployDir:  ".ftl",
		SQLRootDir: sqlRootDir,
	}.FillDefaultsAndValidate(moduleconfig.CustomDefaults{})
	assert.NoError(t, err)
	return mc.Abs()
}
