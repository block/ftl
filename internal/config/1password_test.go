//go:build 1password

// 1password needs to be running and will temporarily make a "ftl-test" vault.
//
// These tests will clean up before and after itself.

package config_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/exec"
)

const vault = "ftl-test"

func Test1PasswordProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 1Password tests when -short")
	}
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.TODO())

	vaultID := createVault(t, ctx)

	onePassword, err := config.NewOnePasswordProvider(vaultID, "unittest")
	assert.NoError(t, err)
	provider := config.NewCacheDecorator(ctx, onePassword)

	testConfig(t, ctx, provider)
}

func createVault(t *testing.T, ctx context.Context) string {
	args := []string{
		"vault", "create", vault,
		"--format", "json",
	}
	output, err := exec.Capture(ctx, ".", "op", args...)
	assert.NoError(t, err, "%s", output)

	t.Cleanup(func() {
		args := []string{"vault", "delete", vault}
		_, err := exec.Capture(ctx, ".", "op", args...)
		assert.NoError(t, err, "failed to delete vault")
	})

	var vault struct {
		ID string `json:"id"`
	}
	err = json.Unmarshal(output, &vault)
	assert.NoError(t, err, "failed to decode 1Password create vault response")
	return vault.ID
}
