package schema

type DeploymentState int

const (
	DeploymentStateUnspecified DeploymentState = iota
	DeploymentStateProvisioning
	DeploymentStateReady
	DeploymentStateCanary
	DeploymentStateCanonical
	DeploymentStateDraining
	DeploymentStateDeProvisioning
	DeploymentStateDeleted
	DeploymentStateFailed
)

//sumtype:decl
//protobuf:export
type Runtime interface {
	runtimeElement()
}

// ModuleRuntime is runtime configuration for a module that can be dynamically updated.
type ModuleRuntime struct {
	Base       ModuleRuntimeBase        `protobuf:"1"` // Base is always present.
	Scaling    *ModuleRuntimeScaling    `protobuf:"2,optional"`
	Deployment *ModuleRuntimeDeployment `protobuf:"3,optional"`
	Runner     *ModuleRuntimeRunner     `protobuf:"4,optional"`
}

func (m *ModuleRuntime) GetScaling() *ModuleRuntimeScaling {
	if m == nil {
		return nil
	}
	return m.Scaling
}

func (m *ModuleRuntime) GetDeployment() *ModuleRuntimeDeployment {
	if m == nil {
		return nil
	}
	return m.Deployment
}
func (m *ModuleRuntime) GetRunner() *ModuleRuntimeRunner {
	if m == nil {
		return nil
	}
	return m.Runner
}
func (m *ModuleRuntime) ModRunner() *ModuleRuntimeRunner {
	if m.Runner == nil {
		m.Runner = &ModuleRuntimeRunner{}
	}
	return m.Runner
}

func (m *ModuleRuntime) ModDeployment() *ModuleRuntimeDeployment {
	if m.Deployment == nil {
		m.Deployment = &ModuleRuntimeDeployment{}
	}
	return m.Deployment
}

func (m *ModuleRuntime) ModScaling() *ModuleRuntimeScaling {
	if m.Scaling == nil {
		m.Scaling = &ModuleRuntimeScaling{}
	}
	return m.Scaling
}
