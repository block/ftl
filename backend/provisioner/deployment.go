package provisioner

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type TaskState string

const (
	TaskStatePending TaskState = ""
	TaskStateRunning TaskState = "running"
	TaskStateDone    TaskState = "done"
	TaskStateFailed  TaskState = "failed"
)

// Task is a unit of work for a deployment
type Task struct {
	binding *ProvisionerBinding
	module  string
	state   TaskState

	deployment *Deployment

	// set if the task is currently running
	runningToken string
}

func (t *Task) Start(ctx context.Context, eventSource *schemaeventsource.EventSource) error {
	if t.state != TaskStatePending {
		return fmt.Errorf("task state is not pending: %s", t.state)
	}
	t.state = TaskStateRunning

	var previous *schemapb.Module
	if t.deployment.Previous != nil {
		previous = t.deployment.Previous.ToProto()
	}
	module, err := t.currentModuleState(eventSource)
	if err != nil {
		return fmt.Errorf("error getting module: %w", err)
	}

	resp, err := t.binding.Provisioner.Provision(ctx, connect.NewRequest(&provisioner.ProvisionRequest{
		DesiredModule: module.ToProto(),
		// TODO: We need a proper cluster specific ID here
		FtlClusterId:   "ftl",
		PreviousModule: previous,
		Kinds:          slices.Map(t.binding.Types, func(x schema.ResourceType) string { return string(x) }),
	}))
	if err != nil {
		t.state = TaskStateFailed
		return fmt.Errorf("error provisioning resources: %w", err)
	}
	t.runningToken = resp.Msg.ProvisioningToken

	return nil
}

func (t *Task) Progress(ctx context.Context, eventSource *schemaeventsource.EventSource) error {
	if t.state != TaskStateRunning {
		return fmt.Errorf("task state is not running: %s", t.state)
	}

	retry := backoff.Backoff{
		Min: 50 * time.Millisecond,
		Max: 30 * time.Second,
	}

	for {

		module, err := t.currentModuleState(eventSource)
		if err != nil {
			return fmt.Errorf("error getting module: %w", err)
		}
		resp, err := t.binding.Provisioner.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
			ProvisioningToken: t.runningToken,
			DesiredModule:     module.ToProto(),
		}))
		if err != nil {
			t.state = TaskStateFailed
			return fmt.Errorf("error getting state: %w", err)
		}
		if succ, ok := resp.Msg.Status.(*provisioner.StatusResponse_Success); ok {
			t.state = TaskStateDone
			events := succ.Success.Events

			for _, eventpb := range events {
				err = t.deployment.EventHandler(eventpb)
				if err != nil {
					return fmt.Errorf("schema server failed to handle provisioning event: %w", err)
				}
			}
			return nil

		}
		time.Sleep(retry.Duration())
	}
}

func (t *Task) currentModuleState(eventSource *schemaeventsource.EventSource) (*schema.Module, error) {
	changeset := eventSource.ActiveChangeset()
	if !changeset.Ok() {
		return nil, fmt.Errorf("no active changeset")
	}
	cs := changeset.MustGet()
	var module *schema.Module
	for _, m := range cs.Modules {
		if m.Name == t.module {
			module = m
			break
		}
	}
	return module, nil
}

// Deployment is a single deployment of resources for a single module
type Deployment struct {
	Tasks []*Task

	Previous     *schema.Module
	EventHandler func(event *schemapb.Event) error
}

// next running or pending task. Nil if all tasks are done.
func (d *Deployment) next() optional.Option[*Task] {
	for _, t := range d.Tasks {
		if t.state == TaskStatePending || t.state == TaskStateRunning || t.state == TaskStateFailed {
			return optional.Some(t)
		}
	}
	return optional.None[*Task]()
}

// Progress the deployment. Returns true if there are still tasks running or pending.
func (d *Deployment) Progress(ctx context.Context, source *schemaeventsource.EventSource) (bool, error) {
	logger := log.FromContext(ctx)

	next, ok := d.next().Get()
	if !ok {
		return false, nil
	}

	if next.state == TaskStatePending {
		logger.Debugf("Starting task %s: %s", next.module, next.binding.ID)
		err := next.Start(ctx, source)
		if err != nil {
			return true, err
		}
	}
	if next.state != TaskStateDone {
		logger.Tracef("Progressing task %s: %s", next.module, next.binding.ID)
		err := next.Progress(ctx, source)
		if err != nil {
			return true, err
		}
	}
	logger.Debugf("Finished task %s: %s", next.module, next.binding.ID)
	return d.next().Ok(), nil
}

type DeploymentState struct {
	Pending []*Task
	Running *Task
	Failed  *Task
	Done    []*Task
}

func (d *Deployment) State() *DeploymentState {
	result := &DeploymentState{}
	for _, t := range d.Tasks {
		switch t.state {
		case TaskStatePending:
			result.Pending = append(result.Pending, t)
		case TaskStateRunning:
			result.Running = t
		case TaskStateFailed:
			result.Failed = t
		case TaskStateDone:
			result.Done = append(result.Done, t)
		}
	}
	return result
}
