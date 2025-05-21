package proxy

import (
	"context"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/block/ftl/backend/controller/observability"
	ftllease "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	ftlleaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/timelineclient"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftlv1connect.ControllerServiceHandler = &Service{}

type moduleVerbService struct {
	client     ftlv1connect.VerbServiceClient
	deployment key.Deployment
	uri        string
}

type Service struct {
	controllerDeploymentService ftlv1connect.ControllerServiceClient
	controllerLeaseService      ftlleaseconnect.LeaseServiceClient
	moduleVerbService           *xsync.MapOf[string, moduleVerbService]
	timelineClient              *timelineclient.Client
	queryService                *query.Service
	module                      *schema.Module
	localModuleName             string
	bindAddress                 string
	localDeployment             key.Deployment
	localRunner                 bool
}

func New(controllerModuleService ftlv1connect.ControllerServiceClient,
	leaseClient ftlleaseconnect.LeaseServiceClient,
	timelineClient *timelineclient.Client,
	queryService *query.Service,
	bindAddress string,
	module *schema.Module,
	localDeployment key.Deployment,
	localRunners bool) *Service {
	proxy := &Service{
		controllerDeploymentService: controllerModuleService,
		controllerLeaseService:      leaseClient,
		moduleVerbService:           xsync.NewMapOf[string, moduleVerbService](),
		timelineClient:              timelineClient,
		queryService:                queryService,
		localModuleName:             localDeployment.Payload.Module,
		bindAddress:                 bindAddress,
		module:                      module,
		localDeployment:             localDeployment,
		localRunner:                 localRunners,
	}
	return proxy
}

func (r *Service) GetDeploymentContext(ctx context.Context, c *connect.Request[ftlv1.GetDeploymentContextRequest], c2 *connect.ServerStream[ftlv1.GetDeploymentContextResponse]) error {
	moduleContext, err := r.controllerDeploymentService.GetDeploymentContext(ctx, connect.NewRequest(c.Msg))
	logger := log.FromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get module context")
	}
	for {
		rcv := moduleContext.Receive()

		if rcv {
			logger.Debugf("Received DeploymentContext from module: %v", moduleContext.Msg())
			for _, route := range moduleContext.Msg().Routes {
				logger.Debugf("Adding proxy route: %s -> %s", route.Deployment, route.Uri)

				deployment, err := key.ParseDeploymentKey(route.Deployment)
				if err != nil {
					return errors.Wrap(err, "failed to parse deployment key")
				}
				module := deployment.Payload.Module
				if existing, ok := r.moduleVerbService.Load(module); !ok || existing.deployment.String() != deployment.String() {
					r.moduleVerbService.Store(module, moduleVerbService{
						client:     rpc.Dial(ftlv1connect.NewVerbServiceClient, route.Uri, log.Error),
						deployment: deployment,
						uri:        route.Uri,
					})
				}
			}
			logger.Debugf("Adding localhost route: %s -> %s", r.localDeployment, r.bindAddress)
			r.moduleVerbService.Store(r.localModuleName, moduleVerbService{
				client:     rpc.Dial(ftlv1connect.NewVerbServiceClient, r.bindAddress, log.Error),
				deployment: r.localDeployment,
				uri:        r.bindAddress,
			})
			err := c2.Send(moduleContext.Msg())
			if err != nil {
				return errors.Wrap(err, "failed to send message")
			}
		} else if moduleContext.Err() != nil {
			return errors.Wrap(moduleContext.Err(), "failed to receive message")
		}
	}

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
	// TODO: timeline code starts here
	if verb, ok := r.getLocalQueryVerb(req.Msg.Verb).Get(); ok {
		return r.queryService.Call(ctx, verb, req.Msg.Body)
	}
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
		DeploymentKey: verbService.deployment,
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

func (r *Service) getLocalQueryVerb(verb *schemapb.Ref) optional.Option[*schema.Verb] {
	if verb.Module != r.module.Name {
		return optional.None[*schema.Verb]()
	}

	decl, ok := slices.Find(r.module.Decls, func(d schema.Decl) bool {
		return d.GetName() == verb.Name
	})
	if !ok {
		return optional.None[*schema.Verb]()
	}
	verbDecl, ok := decl.(*schema.Verb)
	if !ok {
		return optional.None[*schema.Verb]()
	}
	if !verbDecl.IsQuery() {
		return optional.None[*schema.Verb]()
	}
	return optional.Some(verbDecl)
}

func (r *Service) ProcessList(ctx context.Context, c *connect.Request[ftlv1.ProcessListRequest]) (*connect.Response[ftlv1.ProcessListResponse], error) {
	return nil, errors.Errorf("should never be called, this is temp refactoring debt")
}

func (r *Service) Status(ctx context.Context, c *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	return nil, errors.Errorf("should never be called, this is temp refactoring debt")
}

func (r *Service) RegisterRunner(ctx context.Context, c *connect.ClientStream[ftlv1.RegisterRunnerRequest]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {
	return nil, errors.Errorf("should never be called, this is temp refactoring debt")
}
