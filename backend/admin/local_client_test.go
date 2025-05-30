//go:build integration

package admin

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	configuration "github.com/block/ftl/internal/config"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/projectconfig"
)

func getDiskSchema(t testing.TB, ctx context.Context) (*schema.Schema, error) {
	t.Helper()
	projConfig, err := projectconfig.Load(ctx, None[string]())
	assert.NoError(t, err)
	dsr := newDiskSchemaRetriever(projConfig)
	return errors.WithStack2(dsr.GetSchema(ctx))
}

func TestDiskSchemaRetrieverWithBuildArtefact(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule("dischema"),
		in.Build("dischema"),
		func(t testing.TB, ic in.TestContext) {
			sch, err := getDiskSchema(t, ic.Context)
			assert.NoError(t, err)

			module, ok := sch.Module("dr", "dischema").Get()
			assert.Equal(t, ok, true)
			assert.Equal(t, "dischema", module.Name)
		},
	)
}

func TestDiskSchemaRetrieverWithNoSchema(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("ftl-project-dr.toml"),
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule("dischema"),
		func(t testing.TB, ic in.TestContext) {
			_, err := getDiskSchema(t, ic.Context)
			assert.Error(t, err)
		},
	)
}

func TestAdminNoValidationWithNoSchema(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm := configuration.NewMemoryProvider[configuration.Configuration]()
	sm := configuration.NewMemoryProvider[configuration.Secrets]()

	projConfig, err := projectconfig.Load(ctx, None[string]())
	assert.NoError(t, err)

	dsr := newDiskSchemaRetriever(projConfig)
	_, err = dsr.GetSchema(ctx)
	assert.Error(t, err)

	admin := NewEnvironmentClient(cm, sm, dsr)
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
