package plugin_test

import (
	"context"
	"os/exec" //nolint:depguard
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestPlugin(t *testing.T) {
	if testing.Short() {
		t.Skip("not short")
	}
	dir := t.TempDir()
	ctx := t.Context()
	run(ctx, t, ".", "go", "build", "-o", filepath.Join(dir, "host"), "./testdata/host.go")
	run(ctx, t, ".", "go", "build", "-o", filepath.Join(dir, "plugin"), "./testdata/plugin.go")
	ctx, done := context.WithTimeout(ctx, time.Second*2)
	defer done()
	run(ctx, t, dir, "./host")
}

func run(ctx context.Context, t *testing.T, dir, command string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "%s", out)
}
