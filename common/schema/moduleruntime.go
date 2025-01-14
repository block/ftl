package schema

import (
	"time"

	"github.com/block/ftl/internal/key"
)

// ModuleRuntime is runtime configuration for a module that can be dynamically updated.
type ModuleRuntime struct {
	Base       ModuleRuntimeBase        `protobuf:"1"` // Base is always present.
	Scaling    *ModuleRuntimeScaling    `protobuf:"2,optional"`
	Deployment *ModuleRuntimeDeployment `protobuf:"3,optional"`
}

// ApplyEvent applies a ModuleRuntimeEvent to the ModuleRuntime.
func (m *ModuleRuntime) ApplyEvent(event ModuleRuntimeEvent) {
	switch event := event.(type) {
	case *ModuleRuntimeBase:
		m.Base = *event
	case *ModuleRuntimeScaling:
		m.Scaling = event
	case *ModuleRuntimeDeployment:
		m.Deployment = event
	}
}

//sumtype:decl
//protobuf:export
type ModuleRuntimeEvent interface {
	RuntimeEvent

	moduleRuntime()
}

//protobuf:1
//protobuf:1 RuntimeEvent
type ModuleRuntimeBase struct {
	CreateTime time.Time `protobuf:"1"`
	Language   string    `protobuf:"2"`
	OS         string    `protobuf:"3,optional"`
	Arch       string    `protobuf:"4,optional"`
	// Image is the name of the runner image. Defaults to "ftl0/ftl-runner".
	// Must not include a tag, as FTL's version will be used as the tag.
	Image string `protobuf:"5,optional"`
}

func (ModuleRuntimeBase) moduleRuntime() {}

func (m *ModuleRuntimeBase) runtimeEvent() {}

//protobuf:2
//protobuf:2 RuntimeEvent
type ModuleRuntimeScaling struct {
	MinReplicas int32 `protobuf:"1"`
}

func (*ModuleRuntimeScaling) moduleRuntime() {}

func (m *ModuleRuntimeScaling) runtimeEvent() {}

//protobuf:3
//protobuf:3 RuntimeEvent
type ModuleRuntimeDeployment struct {
	// Endpoint is the endpoint of the deployed module.
	Endpoint      string         `protobuf:"1"`
	DeploymentKey key.Deployment `protobuf:"2"`
	CreatedAt     time.Time      `protobuf:"3"`
	ActivatedAt   time.Time      `protobuf:"4"`
}

func (m *ModuleRuntimeDeployment) moduleRuntime() {}

func (m *ModuleRuntimeDeployment) runtimeEvent() {}

func (m *ModuleRuntime) GetScaling() *ModuleRuntimeScaling {
	if m == nil {
		return nil
	}
	return m.Scaling
}

func (m *ModuleRuntimeScaling) GetMinReplicas() int32 {
	if m == nil {
		return 0
	}
	return m.MinReplicas
}

func (m *ModuleRuntime) GetDeployment() *ModuleRuntimeDeployment {
	if m == nil {
		return nil
	}
	return m.Deployment
}

func (m *ModuleRuntime) ModDeployment() *ModuleRuntimeDeployment {
	if m.Deployment == nil {
		m.Deployment = &ModuleRuntimeDeployment{}
	}
	return m.Deployment
}

func (m *ModuleRuntimeDeployment) GetCreatedAt() time.Time {
	if m == nil {
		return time.Time{}
	}
	return m.CreatedAt
}

func (m *ModuleRuntimeDeployment) GetDeploymentKey() key.Deployment {
	if m == nil {
		return key.Deployment{}
	}
	return m.DeploymentKey
}
