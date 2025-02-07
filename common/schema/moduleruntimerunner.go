package schema

//protobuf:3
type ModuleRuntimeRunner struct {
	// Endpoint is the endpoint of the deployed module.
	Endpoint string `protobuf:"1"`
}

var _ Runtime = (*ModuleRuntimeRunner)(nil)

func (m *ModuleRuntimeRunner) runtimeElement() {}

func (m *ModuleRuntimeRunner) GetEndpoint() string {
	if m == nil {
		return ""
	}
	return m.Endpoint

}
