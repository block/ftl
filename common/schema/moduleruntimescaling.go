package schema

//protobuf:2
type ModuleRuntimeScaling struct {
	MinReplicas int32 `protobuf:"1"`
}

var _ Runtime = (*ModuleRuntimeScaling)(nil)

func (m *ModuleRuntimeScaling) runtimeElement() {}

func (m *ModuleRuntimeScaling) GetMinReplicas() int32 {
	if m == nil {
		return 0
	}
	return m.MinReplicas
}
