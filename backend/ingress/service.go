package ingress

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/atomic"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/internal/cors"
	ftlhttp "github.com/block/ftl/internal/http"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type Config struct {
	AllowOrigins     []string `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_INGRESS_ALLOW_ORIGIN"`
	AllowHeaders     []string `help:"Allow these headers in CORS requests. (Requires AllowOrigins)" env:"FTL_INGRESS_ALLOW_HEADERS"`
	IngressURLPrefix string   `help:"URL prefix for ingress endpoints." env:"FTL_INGRESS_URL_PREFIX" default:""`
}

func (c *Config) Validate() error {
	if len(c.AllowHeaders) > 0 && len(c.AllowOrigins) == 0 {
		return errors.Errorf("AllowOrigins must be set when AllowHeaders is used")
	}
	return nil
}

type service struct {
	// Complete schema synchronised from the database.
	view           *atomic.Value[materialisedView]
	client         routing.CallClient
	timelineClient *timelineclient.Client
	routeTable     *routing.RouteTable
	urlPrefix      string
}

// Start the HTTP ingress service. Blocks until the context is cancelled.
func Start(ctx context.Context, bind *url.URL, config Config, eventSource *schemaeventsource.EventSource, client routing.CallClient, timelineClient *timelineclient.Client) error {
	logger := log.FromContext(ctx).Scope("http-ingress")
	ctx = log.ContextWithLogger(ctx, logger)
	svc := &service{
		view:           syncView(ctx, eventSource),
		client:         client,
		timelineClient: timelineClient,
		routeTable:     routing.New(ctx, eventSource),
		urlPrefix:      config.IngressURLPrefix,
	}

	ingressHandler := otelhttp.NewHandler(http.Handler(svc), "ftl.ingress")
	if len(config.AllowOrigins) > 0 {
		ingressHandler = cors.Middleware(
			config.AllowOrigins,
			config.AllowHeaders,
			ingressHandler,
		)
	}

	// Start the HTTP server
	logger.Infof("HTTP ingress server listening on: %s", bind) //nolint
	err := ftlhttp.Serve(ctx, bind, ingressHandler)
	if err != nil {
		return errors.Wrap(err, "ingress service stopped")
	}
	return nil
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if s.urlPrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, s.urlPrefix)
	}

	start := time.Now()
	method := strings.ToLower(r.Method)

	requestKey := key.NewRequestKey(key.OriginIngress, fmt.Sprintf("%s %s", method, r.URL.Path))

	state := s.view.Load()
	routes := state.routes[r.Method]
	if len(routes) == 0 {
		http.NotFound(w, r)
		metrics.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), start, optional.Some("route not found in dal"))
		return
	}
	s.handleHTTP(start, state.schema.WithBuiltins(), requestKey, routes, w, r, s.client)
}
