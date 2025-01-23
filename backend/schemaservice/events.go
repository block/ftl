package schemaservice

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

type SchemaEvent interface {
	Handle(view SchemaState) (SchemaState, error)
}

var _ SchemaEvent = (*DeploymentCreatedEvent)(nil)
var _ SchemaEvent = (*DeploymentActivatedEvent)(nil)
var _ SchemaEvent = (*DeploymentDeactivatedEvent)(nil)
var _ SchemaEvent = (*DeploymentSchemaUpdatedEvent)(nil)
var _ SchemaEvent = (*DeploymentReplicasUpdatedEvent)(nil)

type DeploymentCreatedEvent struct {
	Key    key.Deployment
	Schema *schema.Module
}

func (r *DeploymentCreatedEvent) Handle(t SchemaState) (SchemaState, error) {
	if existing := t.deployments[r.Key]; existing != nil {
		return t, nil
	}
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
	existing.ModRuntime().ModScaling().MinReplicas = int32(r.Replicas)
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
	existing.ModRuntime().ModDeployment().ActivatedAt = optional.Some(r.ActivatedAt)
	existing.ModRuntime().ModScaling().MinReplicas = int32(r.MinReplicas)
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
	existing.ModRuntime().ModScaling().MinReplicas = 0
	delete(t.activeDeployments, r.Key)
	return t, nil
}
