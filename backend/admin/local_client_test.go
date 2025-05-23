//go:build integration

package admin

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	cf "github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/configuration/routers"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
)

func getDiskSchema(t testing.TB, ctx context.Context) (*schema.Schema, error) {
	t.Helper()
	projConfig, err := projectconfig.Load(ctx, optional.None[string]())
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
	config := tempConfigPath(t, "testdata/ftl-project.toml", "admin")
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm, err := manager.New(ctx, routers.ProjectConfig[cf.Configuration]{Config: config}, providers.NewInline[cf.Configuration]())
	assert.NoError(t, err)

	sm, err := manager.New(ctx, routers.ProjectConfig[cf.Secrets]{Config: config}, providers.NewInline[cf.Secrets]())
	assert.NoError(t, err)

	projConfig, err := projectconfig.Load(ctx, optional.None[string]())
	assert.NoError(t, err)

	dsr := newDiskSchemaRetriever(projConfig)
	_, err = dsr.GetSchema(ctx)
	assert.Error(t, err)

	admin := NewEnvironmentClient(cm, sm, dsr)
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
