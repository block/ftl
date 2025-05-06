package routing

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/schema/builder"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func TestRouting(t *testing.T) {
	events := schemaeventsource.NewUnattached()
	assert.NoError(t, events.PublishModuleForTest(
		builder.Module("time").
			Runtime(
				&schema.ModuleRuntime{
					Deployment: &schema.ModuleRuntimeDeployment{
						DeploymentKey: deploymentKey(t, "dpl-default-time-sjkfislfjslfas"),
					},
					Runner: &schema.ModuleRuntimeRunner{
						Endpoint: "http://time.ftl",
					},
				},
			).
			MustBuild(),
	))

	rt := New(log.ContextWithNewDefaultLogger(context.TODO()), events)
	current := rt.Current()
	assert.Equal(t, optional.Ptr(must.Get(url.Parse("http://time.ftl"))), current.GetForModule("time"))
	assert.Equal(t, optional.None[url.URL](), current.GetForModule("echo"))

	assert.NoError(t, events.PublishModuleForTest(
		builder.Module("echo").
			Runtime(&schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{
					DeploymentKey: deploymentKey(t, "dpl-default-echo-sjkfiaslfjslfs"),
				},
				Runner: &schema.ModuleRuntimeRunner{
					Endpoint: "http://echo.ftl",
				},
			}).
			MustBuild(),
	))

	time.Sleep(time.Millisecond * 250)
	current = rt.Current()
	assert.Equal(t, optional.Ptr(must.Get(url.Parse("http://echo.ftl"))), current.GetForModule("echo"))
}

func deploymentKey(t *testing.T, deploymentKey string) key.Deployment {
	t.Helper()
	key, err := key.ParseDeploymentKey(deploymentKey)
	assert.NoError(t, err)
	return key
}
