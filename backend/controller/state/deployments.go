package state

import (
	"fmt"
	"time"

	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func (r *SchemaState) GetDeployment(deployment key.Deployment) (*schema.Module, error) {
	d, ok := r.deployments[deployment]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	return r.deployments
}

func (r *SchemaState) GetActiveDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for key, active := range r.activeDeployments {
		if active {
			deployments[key] = r.deployments[key]
		}
	}
	return deployments
}

func (r *SchemaState) GetActiveDeploymentSchemas() []*schema.Module {
	return maps.Values(r.GetActiveDeployments())
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
}

func (r *DeploymentCreatedEvent) Handle(t SchemaState) (SchemaState, error) {
	if existing := t.deployments[r.Key]; existing != nil {
		return t, nil
	}
	r.Schema.ModRuntime().ModDeployment().CreatedAt = r.CreatedAt
	t.deployments[r.Key] = r.Schema
	return t, nil
}

type DeploymentSchemaUpdatedEvent struct {
	Key    key.Deployment
	Schema *schema.Module
}

func (r *DeploymentSchemaUpdatedEvent) Handle(t SchemaState) (SchemaState, error) {
	_, ok := t.deployments[r.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", r.Key)
	}
	t.deployments[r.Key] = r.Schema
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
	setMinReplicas(existing, r.Replicas)
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
	existing.ModRuntime().ModDeployment().ActivatedAt = r.ActivatedAt
	setMinReplicas(existing, r.MinReplicas)
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
	setMinReplicas(existing, 0)
	delete(t.activeDeployments, r.Key)
	return t, nil
}

func setMinReplicas(s *schema.Module, minReplicas int) {
	if s.Runtime == nil {
		s.Runtime = &schema.ModuleRuntime{}
	}
	if s.Runtime.Scaling == nil {
		s.Runtime.Scaling = &schema.ModuleRuntimeScaling{}
	}
	s.Runtime.Scaling.MinReplicas = int32(minReplicas)
}
