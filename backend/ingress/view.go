package ingress

import (
	"context"

	"github.com/alecthomas/atomic"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

// Synchronise schema changes into a materialised view of the ingress routing table.
func syncView(ctx context.Context, schemaEventSource *schemaeventsource.EventSource) *atomic.Value[materialisedView] {
	logger := log.FromContext(ctx).Scope("http-ingress")
	out := atomic.New(materialisedView{
		routes: map[string][]ingressRoute{},
		schema: &schema.Schema{},
	})
	logger.Debugf("Starting routing sync from schema")
	go func() {
		for range channels.IterContext(ctx, schemaEventSource.Subscribe(ctx)) {
			state := extractIngressRoutingEntries(schemaEventSource.CanonicalView())
			out.Store(state)
		}
	}()
	return out
}

type materialisedView struct {
	routes map[string][]ingressRoute
	schema *schema.Schema
}

type ingressRoute struct {
	path   string
	module string
	verb   string
	method string
}

func extractIngressRoutingEntries(sch *schema.Schema) materialisedView {
	out := materialisedView{
		schema: sch,
		routes: make(map[string][]ingressRoute, len(sch.InternalModules())*2),
	}
	for _, module := range sch.InternalModules() {
		for _, decl := range module.Decls {
			if verb, ok := decl.(*schema.Verb); ok {
				for _, metadata := range verb.Metadata {
					if ingress, ok := metadata.(*schema.MetadataIngress); ok {
						out.routes[ingress.Method] = append(out.routes[ingress.Method], ingressRoute{
							verb:   verb.Name,
							method: ingress.Method,
							path:   ingress.PathString(),
							module: module.Name,
						})
					}
				}
			}
		}
	}
	return out
}
