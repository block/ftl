package state

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/model"
)

type Deployment struct {
	Key         model.DeploymentKey
	Module      string
	Schema      *schema.Module
	MinReplicas int
	CreatedAt   time.Time
	ActivatedAt optional.Option[time.Time]
	Artefacts   map[string]*DeploymentArtefact
	Language    string
}

func (r *SchemaState) GetDeployment(deployment model.DeploymentKey) (*Deployment, error) {
	d, ok := r.deployments[deployment]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

func (r *SchemaState) GetDeployments() map[model.DeploymentKey]*Deployment {
	return r.deployments
}

func (r *SchemaState) GetActiveDeployments() map[model.DeploymentKey]*Deployment {
	deployments := map[model.DeploymentKey]*Deployment{}
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
	Key       model.DeploymentKey
	CreatedAt time.Time
	Module    string
	Schema    *schema.Module
	Artefacts []*DeploymentArtefact
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
		Module:    r.Module,
		Artefacts: map[string]*DeploymentArtefact{},
		Language:  r.Language,
	}
	for _, a := range r.Artefacts {
		n.Artefacts[a.Digest.String()] = &DeploymentArtefact{
			Digest:     a.Digest,
			Path:       a.Path,
			Executable: a.Executable,
		}
	}
	t.deployments[r.Key] = &n
	return t, nil
}

type DeploymentSchemaUpdatedEvent struct {
	Key    model.DeploymentKey
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
	Key      model.DeploymentKey
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
	Key         model.DeploymentKey
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
	Key           model.DeploymentKey
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
