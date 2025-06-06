package routing

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/block/ftl/backend/controller/observability"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

var _ CallClient = (*VerbCallRouter)(nil)

type CallClient interface {
	Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

// VerbCallRouter managed clients for the routing service, so calls to a given module can be routed to the correct instance.
type VerbCallRouter struct {
	routingTable   *RouteTable
	moduleClients  *xsync.MapOf[string, optional.Option[ftlv1connect.VerbServiceClient]]
	timelineClient timelineclient.Publisher
}

func (s *VerbCallRouter) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	start := time.Now()

	client, deployment, ok := s.LookupClient(ctx, req.Msg.Verb.Module)
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment for module"))
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("deployment not found")))
	}

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get callers"))
		return nil, errors.Wrap(err, "could not get callers from headers")
	}

	requestKey, ok, err := headers.GetRequestKey(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get request key"))
		return nil, errors.Wrap(err, "could not process headers for request key")
	} else if !ok {
		requestKey = key.NewRequestKey(key.OriginIngress, "grpc")
		headers.SetRequestKey(req.Header(), requestKey)
	}

	verb, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get verb ref"))
		return nil, errors.Wrap(err, "could not get verb ref")
	}
	callEvent := &timelineclient.Call{
		DeploymentKey: deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      verb,
		Callers:       callers,
		Request:       req.Msg,
	}

	originalResp, err := client.Call(ctx, req)
	if err != nil {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		return nil, errors.Wrapf(err, "failed to call %s", callEvent.DestVerb)
	}
	resp := connect.NewResponse(originalResp.Msg)
	callEvent.Response = result.Ok(resp.Msg)
	s.timelineClient.Publish(ctx, callEvent)
	observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	headers.SetRequestKey(resp.Header(), requestKey)
	return resp, nil
}

func NewVerbRouterFromTable(ctx context.Context, routeTable *RouteTable, timelineClient timelineclient.Publisher) *VerbCallRouter {
	svc := &VerbCallRouter{
		routingTable:   routeTable,
		moduleClients:  xsync.NewMapOf[string, optional.Option[ftlv1connect.VerbServiceClient]](),
		timelineClient: timelineClient,
	}
	routeUpdates := svc.routingTable.Subscribe()
	logger := log.FromContext(ctx)
	go func() {
		defer svc.routingTable.Unsubscribe(routeUpdates)
		for module := range channels.IterContext(ctx, routeUpdates) {
			logger.Tracef("Removing client for module %s", module)
			svc.moduleClients.Delete(module)
		}
	}()
	return svc
}
func NewVerbRouter(ctx context.Context, changes *schemaeventsource.EventSource, timelineClient timelineclient.Publisher) *VerbCallRouter {
	return NewVerbRouterFromTable(ctx, New(ctx, changes), timelineClient)
}

func (s *VerbCallRouter) LookupClient(ctx context.Context, module string) (client ftlv1connect.VerbServiceClient, deployment key.Deployment, ok bool) {
	logger := log.FromContext(ctx)
	current := s.routingTable.Current()
	deployment, ok = current.GetDeployment(module).Get()
	if !ok {
		return nil, key.Deployment{}, false
	}

	res, _ := s.moduleClients.LoadOrCompute(module, func() optional.Option[ftlv1connect.VerbServiceClient] {
		route, ok := current.Get(deployment).Get()
		if !ok {
			logger.Debugf("Failed to find route for deployment %s", deployment)
			return optional.None[ftlv1connect.VerbServiceClient]()
		}
		logger.Debugf("Found route for deployment %s: %s", deployment, route.String())
		return optional.Some(rpc.Dial(ftlv1connect.NewVerbServiceClient, route.String(), log.Error))
	})

	if !res.Ok() {
		return nil, key.Deployment{}, false
	}

	return res.MustGet(), deployment, true
}
