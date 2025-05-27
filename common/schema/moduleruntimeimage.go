package schema

//protobuf:4
type ModuleRuntimeImage struct {
	Image string `protobuf:"2"`
}

var _ Runtime = (*ModuleRuntimeImage)(nil)

func (m *ModuleRuntimeImage) runtimeElement() {}

func (m *ModuleRuntimeImage) GetImage() string {
	if m == nil {
		return ""
	}
	return m.Image
}
