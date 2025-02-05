package schema

import (
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

//protobuf:export
type RuntimeElement struct {
	Element    Runtime                 `protobuf:"1"`
	Deployment key.Deployment          `protobuf:"2"`
	Name       optional.Option[string] `protobuf:"3"`
}

func (x *RuntimeElement) ApplyToModule(state *Module) error {
	switch v := x.Element.(type) {
	case *ModuleRuntimeDeployment:
		state.Runtime.Deployment = v
	case *ModuleRuntimeScaling:
		state.Runtime.Scaling = v
	case *ModuleRuntimeRunner:
		state.Runtime.Runner = v
	case *VerbRuntime:
		d, err := findDecl[*Verb](state, x)
		if err != nil {
			return err
		}
		(*d).Runtime = v
	case *TopicRuntime:
		d, err := findDecl[*Topic](state, x)
		if err != nil {
			return err
		}
		(*d).Runtime = v
	case *DatabaseRuntime:
		d, err := findDecl[*Database](state, x)
		if err != nil {
			return err
		}
		(*d).Runtime = v
	}
	return nil
}

func findDecl[T any](module *Module, x *RuntimeElement) (*T, error) {
	name, ok := x.Name.Get()
	if !ok {
		return nil, fmt.Errorf("missing name in element")
	}
	for _, decl := range module.Decls {
		if decl.GetName() == name {
			if d, ok := decl.(T); ok {
				return &d, nil
			}
			return nil, fmt.Errorf("unexpected decl type %T", decl)
		}
	}
	return nil, fmt.Errorf("decl %s not found", name)
}
