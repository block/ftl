package flock

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/log"
)

func TestFlock(t *testing.T) {
	dir := t.TempDir()
	lockfile := filepath.Join(dir, "lock")
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	release, err := Acquire(ctx, lockfile, 0)
	assert.NoError(t, err)

	_, err = Acquire(ctx, lockfile, 0)
	assert.Error(t, err)

	err = release()
	assert.NoError(t, err)

	releaseb, err := Acquire(ctx, lockfile, 0)
	assert.NoError(t, err)
	err = releaseb()
	assert.NoError(t, err)
}
