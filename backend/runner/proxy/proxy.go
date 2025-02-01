package proxy

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/block/ftl/backend/controller/observability"
	ftldeployment "github.com/block/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	ftldeploymentconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/deployment/v1/deploymentpbconnect"
	ftllease "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1"
	ftlleaseconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/lease/v1/leasepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/timelineclient"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftldeploymentconnect.DeploymentServiceHandler = &Service{}

type moduleVerbService struct {
	client     ftlv1connect.VerbServiceClient
	deployment key.Deployment
	uri        string
}

type Service struct {
	controllerDeploymentService ftldeploymentconnect.DeploymentServiceClient
	controllerLeaseService      ftlleaseconnect.LeaseServiceClient
	moduleVerbService           *xsync.MapOf[string, moduleVerbService]
	timelineClient              *timelineclient.Client
	queryService                *query.MultiService
	localModuleName             string
	bindAddress                 string
	localDeployment             key.Deployment
}

func New(controllerModuleService ftldeploymentconnect.DeploymentServiceClient,
	leaseClient ftlleaseconnect.LeaseServiceClient,
	timelineClient *timelineclient.Client,
	queryService *query.MultiService,
	bindAddress string,
	localDeployment key.Deployment) *Service {
	proxy := &Service{
		controllerDeploymentService: controllerModuleService,
		controllerLeaseService:      leaseClient,
		moduleVerbService:           xsync.NewMapOf[string, moduleVerbService](),
		timelineClient:              timelineClient,
		queryService:                queryService,
		localModuleName:             localDeployment.Payload.Module,
		bindAddress:                 bindAddress,
		localDeployment:             localDeployment,
	}
	return proxy
}

func (r *Service) GetDeploymentContext(ctx context.Context, c *connect.Request[ftldeployment.GetDeploymentContextRequest], c2 *connect.ServerStream[ftldeployment.GetDeploymentContextResponse]) error {
	moduleContext, err := r.controllerDeploymentService.GetDeploymentContext(ctx, connect.NewRequest(c.Msg))
	logger := log.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module context: %w", err)
	}
	for {
		rcv := moduleContext.Receive()

		if rcv {
			logger.Debugf("Received DeploymentContext from module: %v", moduleContext.Msg())
			for _, route := range moduleContext.Msg().Routes {
				logger.Debugf("Adding proxy route: %s -> %s", route.Deployment, route.Uri)

				deployment, err := key.ParseDeploymentKey(route.Deployment)
				if err != nil {
					return fmt.Errorf("failed to parse deployment key: %w", err)
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
				return fmt.Errorf("failed to send message: %w", err)
			}
		} else if moduleContext.Err() != nil {
			return fmt.Errorf("failed to receive message: %w", moduleContext.Err())
		}
	}

}

func (r *Service) AcquireLease(ctx context.Context, c *connect.BidiStream[ftllease.AcquireLeaseRequest, ftllease.AcquireLeaseResponse]) error {
	_, err := r.controllerLeaseService.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		return fmt.Errorf("failed to ping lease service: %w", err)
	}
	lease := r.controllerLeaseService.AcquireLease(ctx)
	defer lease.CloseResponse() //nolint:errcheck
	defer lease.CloseRequest()  //nolint:errcheck
	for {
		req, err := c.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		err = lease.Send(req)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		msg, err := lease.Receive()
		if err != nil {
			return connect.NewError(connect.CodeOf(err), fmt.Errorf("lease failed %w", err))
		}
		err = c.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send response message: %w", err)
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
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("proxy failed to route request, deployment not found"))
	}

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get callers"))
		return nil, fmt.Errorf("could not get callers from headers: %w", err)
	}

	requestKey, ok, err := headers.GetRequestKey(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get request key"))
		return nil, fmt.Errorf("could not process headers for request key: %w", err)
	} else if !ok {
		requestKey = key.NewRequestKey(key.OriginIngress, "grpc")
		headers.SetRequestKey(req.Header(), requestKey)
	}

	destVerb, err := schema.RefFromProto(req.Msg.Verb)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get verb ref"))
		return nil, fmt.Errorf("could not get verb ref: %w", err)
	}
	callEvent := &timelineclient.Call{
		DeploymentKey: verbService.deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      destVerb,
		Callers:       callers,
		Request:       req.Msg,
	}

	originalResp, err := verbService.client.Call(ctx, headers.CopyRequestForForwarding(req))
	if err != nil {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		r.timelineClient.Publish(ctx, callEvent)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		return nil, fmt.Errorf("failed to proxy verb to %s: %w", verbService.uri, err)
	}
	resp := connect.NewResponse(originalResp.Msg)
	callEvent.Response = result.Ok(resp.Msg)
	r.timelineClient.Publish(ctx, callEvent)
	observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	return resp, nil
}
