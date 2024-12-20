package schema

// RuntimeEvent is an event modifying a runtime part of the schema.
//
//sumtype:decl
//protobuf:export
type RuntimeEvent interface {
	runtimeEvent()
}
