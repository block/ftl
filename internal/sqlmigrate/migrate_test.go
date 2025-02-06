package sqlmigrate

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	_ "modernc.org/sqlite"

	"github.com/block/ftl/internal/log"
)

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.db"

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	mfs1, err := fs.Sub(migrations, "testdata/migrations0")
	assert.NoError(t, err)

	dsn := must.Get(url.Parse("sqlite://" + path))
	err = Migrate(ctx, dsn, mfs1)
	assert.NoError(t, err)

	var userAndAge string
	db, err := sql.Open("sqlite", "file:"+path)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	err = db.QueryRow(`SELECT name_and_age FROM test`).Scan(&userAndAge)
	assert.NoError(t, err)
	assert.Equal(t, "Alice 30", userAndAge)

	mfs2, err := fs.Sub(migrations, "testdata/migrations1")
	assert.NoError(t, err)
	// Should be idempotent.
	err = Migrate(ctx, dsn, mfs2)
	assert.NoError(t, err)

	type user struct {
		Name string
		Age  int
	}

	var expectedUser user

	err = db.QueryRow(`SELECT name, age FROM test`).Scan(&expectedUser.Name, &expectedUser.Age)
	assert.NotZero(t, err)
	assert.Equal(t, user{"Alice 30", 0}, expectedUser)
}

//go:embed testdata
var migrations embed.FS
