package proxy

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/block/ftl/backend/controller/observability"
	ftllease "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	ftlleaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/timelineclient"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftlv1connect.DeploymentContextServiceHandler = &Service{}

type moduleVerbService struct {
	client ftlv1connect.VerbServiceClient
	uri    string
}

type Service struct {
	deploymentContextProvider deploymentcontext.DeploymentContextProvider
	controllerLeaseService    ftlleaseconnect.LeaseServiceClient
	moduleVerbService         *xsync.MapOf[string, moduleVerbService]
	timelineClient            *timelineclient.Client
	localModuleName           string
	runnerBindAddress         string
	localDeployment           key.Deployment
	localRunner               bool
}

func New(controllerModuleService deploymentcontext.DeploymentContextProvider,
	leaseClient ftlleaseconnect.LeaseServiceClient,
	timelineClient *timelineclient.Client,
	localDeployment key.Deployment,
	localRunners bool) *Service {
	proxy := &Service{
		deploymentContextProvider: controllerModuleService,
		controllerLeaseService:    leaseClient,
		moduleVerbService:         xsync.NewMapOf[string, moduleVerbService](),
		timelineClient:            timelineClient,
		localModuleName:           localDeployment.Payload.Module,
		localDeployment:           localDeployment,
		localRunner:               localRunners,
	}
	return proxy
}

func (r *Service) GetDeploymentContext(ctx context.Context, c *connect.Request[ftlv1.GetDeploymentContextRequest], c2 *connect.ServerStream[ftlv1.GetDeploymentContextResponse]) error {
	logger := log.FromContext(ctx)
	updates := r.deploymentContextProvider(ctx)
	for i := range channels.IterContext[deploymentcontext.DeploymentContext](ctx, updates) {

		logger.Debugf("Received DeploymentContext from module: %v", i.GetModule())
		for module := range i.GetRoutes() {
			if module == r.localModuleName {
				continue
			}
			route := i.GetRoute(module)
			logger.Debugf("Adding proxy route: %s -> %s", module, route)

			if existing, ok := r.moduleVerbService.Load(module); !ok || existing.uri != route {
				r.moduleVerbService.Store(module, moduleVerbService{
					client: rpc.Dial(ftlv1connect.NewVerbServiceClient, route, log.Error),
					uri:    route,
				})
			}
		}
		err := c2.Send(i.ToProto())
		if err != nil {
			return errors.Wrap(err, "failed to send message")
		}
	}
	return nil
}
func (r *Service) SetRunnerAddress(ctx context.Context, address string) {
	logger := log.FromContext(ctx)
	r.runnerBindAddress = address
	logger.Debugf("Adding localhost route: %s -> %s", r.localModuleName, r.runnerBindAddress)
	r.moduleVerbService.Store(r.localModuleName, moduleVerbService{
		client: rpc.Dial(ftlv1connect.NewVerbServiceClient, r.runnerBindAddress, log.Error),
		uri:    r.runnerBindAddress,
	})
}

func (r *Service) AcquireLease(ctx context.Context, c *connect.BidiStream[ftllease.AcquireLeaseRequest, ftllease.AcquireLeaseResponse]) error {
	_, err := r.controllerLeaseService.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		return errors.Wrap(err, "failed to ping lease service")
	}
	lease := r.controllerLeaseService.AcquireLease(ctx)
	defer lease.CloseResponse() //nolint:errcheck
	defer lease.CloseRequest()  //nolint:errcheck
	for {
		req, err := c.Receive()
		if err != nil {
			return errors.Wrap(err, "failed to receive message")
		}
		err = lease.Send(req)
		if err != nil {
			return errors.Wrap(err, "failed to send message")
		}
		msg, err := lease.Receive()
		if err != nil {
			return errors.WithStack(connect.NewError(connect.CodeOf(err), errors.Wrap(err, "lease failed")))
		}
		err = c.Send(msg)
		if err != nil {
			return errors.Wrap(err, "failed to send response message")
		}
	}
}

func (r *Service) Ping(ctx context.Context, c *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (r *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	start := time.Now()
	verbService, ok := r.moduleVerbService.Load(req.Msg.Verb.Module)
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment for module"))
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("proxy failed to route request, deployment not found")))
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

	destVerb, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get verb ref"))
		return nil, errors.Wrap(err, "could not get verb ref")
	}
	callEvent := &timelineclient.Call{
		DeploymentKey: r.localDeployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      destVerb,
		Callers:       callers,
		Request:       req.Msg,
	}

	newRequest := headers.CopyRequestForForwarding(req)
	if r.localRunner {
		// When running locally we emulate spiffe client cert headers that istio would normally add
		// This is so that we can emulate the same behavior as when running in the cloud
		newRequest.Header().Set("x-forwarded-client-cert", "By=spiffe://cluster.local/ns/"+req.Msg.Verb.Module+"/sa/"+req.Msg.Verb.Module+";URI=spiffe://cluster.local/ns/"+r.localModuleName+"/sa/"+r.localModuleName)
	}
	originalResp, err := verbService.client.Call(ctx, newRequest)
	if err != nil {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		r.timelineClient.Publish(ctx, callEvent)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		return nil, errors.Wrapf(err, "failed to proxy verb to %s", verbService.uri)
	}
	resp := connect.NewResponse(originalResp.Msg)
	callEvent.Response = result.Ok(resp.Msg)
	r.timelineClient.Publish(ctx, callEvent)
	observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	return resp, nil
}
