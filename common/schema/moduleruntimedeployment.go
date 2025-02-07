package schema

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

//protobuf:1
type ModuleRuntimeDeployment struct {
	DeploymentKey key.Deployment             `protobuf:"2"`
	CreatedAt     time.Time                  `protobuf:"3"`
	ActivatedAt   optional.Option[time.Time] `protobuf:"4"`
	State         DeploymentState            `protobuf:"5"`
}

var _ Runtime = (*ModuleRuntimeDeployment)(nil)

func (m *ModuleRuntimeDeployment) runtimeElement() {}

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
