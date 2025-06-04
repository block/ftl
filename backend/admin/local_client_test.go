//go:build integration

package admin

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/config"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/profiles"
)

func getDiskSchema(t testing.TB, ctx context.Context, dir string) (*schema.Schema, error) {
	t.Helper()
	project, err := profiles.Open(dir,
		config.NewSecretsRegistry(None[adminpbconnect.AdminServiceClient]()),
		config.NewConfigurationRegistry(None[adminpbconnect.AdminServiceClient]()),
	)
	assert.NoError(t, err)
	dsr := newDiskSchemaRetriever(project.Config())
	return errors.WithStack2(dsr.GetSchema(ctx))
}

func TestDiskSchemaRetrieverWithBuildArtefact(t *testing.T) {
	in.Run(t,
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule("dischema"),
		in.Build("dischema"),
		func(t testing.TB, ic in.TestContext) {
			sch, err := getDiskSchema(t, ic.Context, ic.WorkingDir())
			assert.NoError(t, err)

			module, ok := sch.Module("test", "dischema").Get()
			assert.Equal(t, ok, true, "Couldn't find test.dischema:\n%s", sch)
			assert.Equal(t, "dischema", module.Name)
		},
	)
}

func TestDiskSchemaRetrieverWithNoSchema(t *testing.T) {
	in.Run(t,
		in.WithoutController(),
		in.WithoutTimeline(),
		in.CopyModule("dischema"),
		func(t testing.TB, ic in.TestContext) {
			_, err := getDiskSchema(t, ic.Context, ic.WorkingDir())
			assert.Error(t, err)
		},
	)
}

func TestAdminNoValidationWithNoSchema(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cm := config.NewMemoryProvider[config.Configuration]()
	sm := config.NewMemoryProvider[config.Secrets]()

	project, err := profiles.Open("testdata",
		config.NewSecretsRegistry(None[adminpbconnect.AdminServiceClient]()),
		config.NewConfigurationRegistry(None[adminpbconnect.AdminServiceClient]()),
	)
	assert.NoError(t, err)

	dsr := newDiskSchemaRetriever(project.Config())
	_, err = dsr.GetSchema(ctx)
	assert.Error(t, err)

	admin := NewEnvironmentClient(cm, sm, dsr)
	testSetConfig(t, ctx, admin, "batmobile", "color", "Red", "")
	testSetSecret(t, ctx, admin, "batmobile", "owner", 99, "")
}
