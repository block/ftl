package provisioner

import (
	"context"

	errors "github.com/alecthomas/errors"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
)

// Task is a unit of work for a deployment
type Task struct {
	binding *ProvisionerBinding
	module  string

	deployment *Deployment
}

func (t *Task) Run(ctx context.Context) error {
	var previous *schemapb.Module
	if t.deployment.Previous != nil {
		previous = t.deployment.Previous.ToProto()
	}

	logger := log.FromContext(ctx).
		Deployment(t.deployment.DeploymentState.Runtime.Deployment.DeploymentKey).
		Changeset(t.deployment.Changeset)

	ctx = log.ContextWithLogger(ctx, logger)

	result, err := t.binding.Provisioner.Provision(ctx, &provisioner.ProvisionRequest{
		DesiredModule: t.deployment.DeploymentState.ToProto(),
		// TODO: We need a proper cluster specific ID here
		FtlClusterId:   "ftl",
		PreviousModule: previous,
		Changeset:      t.deployment.Changeset.String(),
		Kinds:          slices.Map(t.binding.Types, func(x schema.ResourceType) string { return string(x) }),
	})
	if err != nil {
		return errors.Wrap(err, "error provisioning resources")
	}
	for _, r := range result {
		element, err := schema.RuntimeElementFromProto(r)
		if err != nil {
			return errors.Wrap(err, "error converting runtime")
		}
		err = element.ApplyToModule(t.deployment.DeploymentState)
		if err != nil {
			return errors.Wrap(err, "error applying runtime")
		}
		err = t.deployment.UpdateHandler(element)
		if err != nil {
			return errors.Wrap(err, "error updating runtime")
		}
	}

	return nil
}

// Deployment is a single deployment of resources for a single module
type Deployment struct {
	Tasks []*Task
	// TODO: Merge runtimes at creation time

	DeploymentState *schema.Module
	Previous        *schema.Module
	Changeset       key.Changeset
	UpdateHandler   func(*schema.RuntimeElement) error
}

// Progress the deployment. Returns true if there are still tasks running or pending.
func (d *Deployment) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	for _, t := range d.Tasks {
		logger.Tracef("Running task %s: %s", t.module, t.binding.ID)
		if err := t.Run(ctx); err != nil {
			return errors.Wrapf(err, "error running task %s", t.module)
		}
	}
	return nil
}

type DeploymentState struct {
	Pending []*Task
	Running *Task
	Failed  *Task
	Done    []*Task
}
