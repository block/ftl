package ingress

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func TestSyncView(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())

	source := schemaeventsource.NewUnattached()
	view := syncView(ctx, source)

	assert.NoError(t, source.PublishModuleForTest(&schema.Module{
		Name: "time",
		Decls: []schema.Decl{
			&schema.Verb{
				Name: "time",
				Metadata: []schema.Metadata{
					&schema.MetadataIngress{
						Type:   "http",
						Method: "GET",
						Path: []schema.IngressPathComponent{
							&schema.IngressPathLiteral{Text: "foo"},
							&schema.IngressPathParameter{Name: "bar"},
						},
					},
				},
			},
		},
	}))

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, materialisedView{
		routes: map[string][]ingressRoute{
			"GET": {
				{path: "/foo/{bar}", module: "time", verb: "time", method: "GET"},
			},
		},
	}, view.Load(), assert.Exclude[*schema.Schema]())
}
