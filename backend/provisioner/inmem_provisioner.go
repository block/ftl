package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

type inMemProvisioningTask struct {
	steps []*inMemProvisioningStep

	events []*schema.RuntimeElement
}

func (t *inMemProvisioningTask) Done() (bool, error) {
	done := true
	for _, step := range t.steps {
		if !step.Done.Load() {
			done = false
		}
		if step.Err != nil {
			return false, step.Err
		}
	}
	return done, nil
}

type inMemProvisioningStep struct {
	Err  error
	Done *atomic.Value[bool]
}

type InMemResourceProvisionerFn func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, resource schema.Provisioned) (*schema.RuntimeElement, error)

// InMemProvisioner for running an in memory provisioner, constructing all resources concurrently
//
// It spawns a separate goroutine for each resource to be provisioned, and
// finishes the task when all resources are provisioned or an error occurs.
type InMemProvisioner struct {
	running        *xsync.MapOf[string, *inMemProvisioningTask]
	handlers       map[schema.ResourceType]InMemResourceProvisionerFn
	removeHandlers map[schema.ResourceType]InMemResourceProvisionerFn
}

func NewEmbeddedProvisioner(handlers map[schema.ResourceType]InMemResourceProvisionerFn, deProvisionHandlers map[schema.ResourceType]InMemResourceProvisionerFn) *InMemProvisioner {
	return &InMemProvisioner{
		running:        xsync.NewMapOf[string, *inMemProvisioningTask](),
		handlers:       handlers,
		removeHandlers: deProvisionHandlers,
	}
}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*InMemProvisioner)(nil)

func (d *InMemProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

type stepCompletedEvent struct {
	step  *inMemProvisioningStep
	event optional.Option[schema.RuntimeElement]
}

func (d *InMemProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	logger := log.FromContext(ctx)
	parsed, err := key.ParseChangesetKey(req.Msg.Changeset)
	if err != nil {
		err = fmt.Errorf("invalid changeset: %w", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var previousModule *schema.Module
	if req.Msg.PreviousModule != nil {
		pm, err := schema.ValidatedModuleFromProto(req.Msg.PreviousModule)
		if err != nil {
			err = fmt.Errorf("invalid previous module: %w", err)
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		previousModule = pm
	}
	desiredModule, err := schema.ValidatedModuleFromProto(req.Msg.DesiredModule)
	if err != nil {
		err = fmt.Errorf("invalid desired module: %w", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	kinds := slices.Map(req.Msg.Kinds, func(k string) schema.ResourceType { return schema.ResourceType(k) })
	previousNodes := schema.GetProvisioned(previousModule)
	desiredNodes := schema.GetProvisioned(desiredModule)

	task := &inMemProvisioningTask{}
	// use chans to safely collect all events before completing each task
	completions := make(chan stepCompletedEvent, 16)

	for id, desired := range desiredNodes {
		previous, prevOk := previousNodes[id]

		for _, resource := range desired.GetProvisioned() {
			if !prevOk || resource.DeploymentSpecific || !resource.IsEqual(previous.GetProvisioned().Get(resource.Kind)) {
				if slices.Contains(kinds, resource.Kind) {
					var handler InMemResourceProvisionerFn
					var ok bool
					if desiredModule.Runtime.Deployment.State == schema.DeploymentStateDeProvisioning {
						handler, ok = d.removeHandlers[resource.Kind]
						if !ok {
							// TODO: should a missing de-provisioner handler be an error?
							continue
						}
					} else {
						handler, ok = d.handlers[resource.Kind]
						if !ok {
							err := fmt.Errorf("unsupported resource type: %s", resource.Kind)
							return nil, connect.NewError(connect.CodeInvalidArgument, err)
						}
					}
					step := &inMemProvisioningStep{Done: atomic.New(false)}
					task.steps = append(task.steps, step)
					go func() {
						event, err := handler(ctx, parsed, desiredModule.Runtime.Deployment.DeploymentKey, desired)
						if err != nil {
							step.Err = err
							completions <- stepCompletedEvent{step: step}
							return
						}
						completions <- stepCompletedEvent{
							step:  step,
							event: optional.Ptr(event),
						}
					}()
				}
			}
		}
	}

	go func() {
		for c := range channels.IterContext(ctx, completions) {
			if e, ok := c.event.Get(); ok {
				task.events = append(task.events, &e)
			}
			c.step.Done.Store(true)
			done, err := task.Done()
			if done || err != nil {
				return
			}
		}
	}()

	token := uuid.New().String()
	logger.Debugf("started a task with token %s", token)
	d.running.Store(token, task)

	return connect.NewResponse(&provisioner.ProvisionResponse{
		ProvisioningToken: token,
		Status:            provisioner.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
	}), nil
}

func (d *InMemProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	logger := log.FromContext(ctx)

	token := req.Msg.ProvisioningToken
	task, ok := d.running.Load(token)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("unknown token: %s", token))
	}
	done, err := task.Done()
	if err != nil {
		logger.Debugf("task with token %s failed with error: %s", token, err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if !done {
		return connect.NewResponse(&provisioner.StatusResponse{
			Status: &provisioner.StatusResponse_Running{},
		}), nil
	}
	logger.Debugf("task with token %s is done", token)

	d.running.Delete(token)

	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Success{
			Success: &provisioner.StatusResponse_ProvisioningSuccess{
				Outputs: slices.Map(task.events, func(e *schema.RuntimeElement) *schemapb.RuntimeElement {
					return e.ToProto()
				}),
			},
		},
	}), nil
}
