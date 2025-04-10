package compile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestBuild(t *testing.T) {
	// set up
	moduleDir, err := filepath.Abs("testdata/echo")
	assert.NoError(t, err)
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	assert.NoError(t, os.RemoveAll(filepath.Join(moduleDir, ".venv")))
	assert.NoError(t, exec.Command(ctx, log.Debug, moduleDir, "uv", "sync").Run())

	t.Run("schema extraction", func(t *testing.T) {
		config := moduleconfig.AbsModuleConfig{
			Dir:    moduleDir,
			Module: "test",
		}
		actual, buildErrors, err := Build(ctx, "", "", config, &schema.Schema{Realms: []*schema.Realm{{Name: "test"}}}, nil, true)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(buildErrors))
		expected := &schema.Module{
			Name: "echo",
			Decls: []schema.Decl{
				&schema.Data{
					Name: "EchoRequest",
					Fields: []*schema.Field{
						{Name: "name", Type: &schema.String{}},
					},
				},
				&schema.Data{
					Name: "EchoResponse",
					Fields: []*schema.Field{
						{Name: "message", Type: &schema.String{}},
					},
				},
				&schema.Verb{Name: "echo",
					Request:  &schema.Ref{Module: "echo", Name: "EchoRequest"},
					Response: &schema.Ref{Module: "echo", Name: "EchoResponse"},
				},
			},
		}
		assert.Equal(t, expected, actual, assert.Exclude[schema.Position]())
	})
}
