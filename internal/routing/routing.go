package routing

import (
	"context"
	"net/url"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type RouteView struct {
	byDeployment       map[string]*url.URL
	moduleToDeployment map[string]key.Deployment
	schema             *schema.Schema
}

type RouteTable struct {
	routes *atomic.Value[RouteView]
	// When the routes for a module change they are published here.
	changeNotification *pubsub.Topic[string]
}

func New(ctx context.Context, changes schemaeventsource.EventSource) *RouteTable {
	r := &RouteTable{
		routes:             atomic.New(extractRoutes(ctx, changes.CanonicalView())),
		changeNotification: pubsub.New[string](),
	}
	go r.run(ctx, changes)
	return r
}

func (r *RouteTable) run(ctx context.Context, changes schemaeventsource.EventSource) {
	for event := range channels.IterContext(ctx, changes.Events()) {
		logger := log.FromContext(ctx)
		logger.Debugf("Received schema event: %T", event)
		old := r.routes.Load()
		routes := extractRoutes(ctx, changes.CanonicalView())
		for module, rd := range old.moduleToDeployment {
			if old.byDeployment[rd.String()] != routes.byDeployment[rd.String()] {
				r.changeNotification.Publish(module)
			}
		}
		for module, rd := range routes.moduleToDeployment {
			// Check for new modules
			if old.byDeployment[rd.String()] == nil {
				r.changeNotification.Publish(module)
			}
		}
		r.routes.Store(routes)
	}
}

// Current returns the current routes.
func (r *RouteTable) Current() RouteView {
	return r.routes.Load()
}

// Get returns the URL for the given deployment or None if it doesn't exist.
func (r RouteView) Get(deployment key.Deployment) optional.Option[url.URL] {
	mod := r.byDeployment[deployment.String()]
	if mod == nil {
		return optional.None[url.URL]()
	}
	return optional.Some(*mod)
}

// GetForModule returns the URL for the given module or None if it doesn't exist.
func (r RouteView) GetForModule(module string) optional.Option[url.URL] {
	dep, ok := r.moduleToDeployment[module]
	if !ok {
		return optional.None[url.URL]()
	}
	return r.Get(dep)
}

// GetDeployment returns the deployment key for the given module or None if it doesn't exist.
func (r RouteView) GetDeployment(module string) optional.Option[key.Deployment] {
	return optional.Zero(r.moduleToDeployment[module])
}

// Schema returns the current schema that the routes are based on.
func (r RouteView) Schema() *schema.Schema {
	return r.schema
}

func (r *RouteTable) Subscribe() chan string {
	return r.changeNotification.Subscribe(nil)
}
func (r *RouteTable) Unsubscribe(s chan string) {
	r.changeNotification.Unsubscribe(s)
}

func extractRoutes(ctx context.Context, sch *schema.Schema) RouteView {
	if sch == nil {
		return RouteView{moduleToDeployment: map[string]key.Deployment{}, byDeployment: map[string]*url.URL{}, schema: &schema.Schema{}}
	}
	logger := log.FromContext(ctx)
	moduleToDeployment := make(map[string]key.Deployment, len(sch.Modules))
	byDeployment := make(map[string]*url.URL, len(sch.Modules))
	for _, module := range sch.Modules {
		if module.GetRuntime() == nil {
			continue
		}
		rt := module.Runtime.Deployment
		if rt.Endpoint == "" {
			logger.Debugf("Skipping route for %s/%s as it is not ready yet", module.Name, rt.DeploymentKey)
			continue
		}
		u, err := url.Parse(rt.Endpoint)
		if err != nil {
			logger.Warnf("Failed to parse endpoint URL for module %q: %v", module.Name, err)
			continue
		}
		logger.Debugf("Adding route for %s/%s: %s", module.Name, rt.DeploymentKey, u)
		moduleToDeployment[module.Name] = rt.DeploymentKey
		byDeployment[rt.DeploymentKey.String()] = u
	}
	return RouteView{moduleToDeployment: moduleToDeployment, byDeployment: byDeployment, schema: sch}
}
