package ingress

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/schema/builder"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

func TestSyncView(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())

	source := schemaeventsource.NewUnattached()
	view := syncView(ctx, source)

	timeModule := builder.Module("time").
		Decl(
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
				Request:  &schema.Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []schema.Type{&schema.Unit{}, &schema.Map{Key: &schema.String{}, Value: &schema.String{}}, &schema.Unit{}}},
				Response: &schema.Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []schema.Type{&schema.Unit{}, &schema.Unit{}}},
			},
		).
		MustBuild()

	assert.NoError(t, source.PublishModuleForTest(timeModule))

	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, materialisedView{
		routes: map[string][]ingressRoute{
			"GET": {
				{path: "/foo/{bar}", module: "time", verb: "time", method: "GET"},
			},
		},
	}, view.Load(), assert.Exclude[*schema.Schema]())
}
