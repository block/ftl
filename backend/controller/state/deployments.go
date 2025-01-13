package state

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
)

type Deployment struct {
	Key         key.Deployment
	Schema      *schema.Module
	MinReplicas int
	CreatedAt   time.Time
	ActivatedAt optional.Option[time.Time]
	Language    string
}

func (r *SchemaState) GetDeployment(deployment key.Deployment) (*Deployment, error) {
	d, ok := r.deployments[deployment]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*Deployment {
	return r.deployments
}

func (r *SchemaState) GetActiveDeployments() map[key.Deployment]*Deployment {
	deployments := map[key.Deployment]*Deployment{}
	for key, active := range r.activeDeployments {
		if active {
			deployments[key] = r.deployments[key]
		}
	}
	return deployments
}

func (r *SchemaState) GetActiveDeploymentSchemas() []*schema.Module {
	rows := r.GetActiveDeployments()
	return slices.Map(maps.Values(rows), func(in *Deployment) *schema.Module { return in.Schema })
}

var _ SchemaEvent = (*DeploymentCreatedEvent)(nil)
var _ SchemaEvent = (*DeploymentActivatedEvent)(nil)
var _ SchemaEvent = (*DeploymentDeactivatedEvent)(nil)
var _ SchemaEvent = (*DeploymentSchemaUpdatedEvent)(nil)
var _ SchemaEvent = (*DeploymentReplicasUpdatedEvent)(nil)

type DeploymentCreatedEvent struct {
	Key       key.Deployment
	CreatedAt time.Time
	Schema    *schema.Module
	Language  string
}

func (r *DeploymentCreatedEvent) Handle(t SchemaState) (SchemaState, error) {
	if existing := t.deployments[r.Key]; existing != nil {
		return t, nil
	}
	n := Deployment{
		Key:       r.Key,
		CreatedAt: r.CreatedAt,
		Schema:    r.Schema,
		Language:  r.Language,
	}
	t.deployments[r.Key] = &n
	return t, nil
}

type DeploymentSchemaUpdatedEvent struct {
	Key    key.Deployment
	Schema *schema.Module
}

func (r *DeploymentSchemaUpdatedEvent) Handle(t SchemaState) (SchemaState, error) {
	existing, ok := t.deployments[r.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)
	}
	existing.Schema = r.Schema
	return t, nil
}

type DeploymentReplicasUpdatedEvent struct {
	Key      key.Deployment
	Replicas int
}

func (r *DeploymentReplicasUpdatedEvent) Handle(t SchemaState) (SchemaState, error) {
	existing, ok := t.deployments[r.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)
	}
	if existing.Schema.Runtime == nil {
		existing.Schema.Runtime = &schema.ModuleRuntime{}
	}
	if existing.Schema.Runtime.Scaling == nil {
		existing.Schema.Runtime.Scaling = &schema.ModuleRuntimeScaling{}
	}
	existing.Schema.Runtime.Scaling.MinReplicas = int32(r.Replicas)
	existing.MinReplicas = r.Replicas
	return t, nil
}

type DeploymentActivatedEvent struct {
	Key         key.Deployment
	ActivatedAt time.Time
	MinReplicas int
}

func (r *DeploymentActivatedEvent) Handle(t SchemaState) (SchemaState, error) {
	existing, ok := t.deployments[r.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)

	}
	existing.ActivatedAt = optional.Some(r.ActivatedAt)
	existing.MinReplicas = r.MinReplicas
	t.activeDeployments[r.Key] = true
	return t, nil
}

type DeploymentDeactivatedEvent struct {
	Key           key.Deployment
	ModuleRemoved bool
}

func (r *DeploymentDeactivatedEvent) Handle(t SchemaState) (SchemaState, error) {
	existing, ok := t.deployments[r.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)

	}
	existing.MinReplicas = 0
	delete(t.activeDeployments, r.Key)
	return t, nil
}
