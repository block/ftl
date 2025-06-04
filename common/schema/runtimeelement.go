package schema

import (
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/key"
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
		if v.DeploymentKey.IsZero() {
			v.DeploymentKey = state.Runtime.Deployment.DeploymentKey
		}
		state.Runtime.Deployment = v
	case *ModuleRuntimeScaling:
		state.Runtime.Scaling = v
	case *ModuleRuntimeRunner:
		state.Runtime.Runner = v
	case *ModuleRuntimeImage:
		state.Runtime.Image = v
	case SubscriptionConnector:
		d, err := findDecl[*Verb](state, x)
		if err != nil {
			return errors.WithStack(err)
		}
		if (*d).Runtime == nil {
			(*d).Runtime = &VerbRuntime{}
		}
		(*d).Runtime.SubscriptionConnector = v
	case *TopicRuntime:
		d, err := findDecl[*Topic](state, x)
		if err != nil {
			return errors.WithStack(err)
		}
		(*d).Runtime = v
	case *DatabaseRuntime:
		d, err := findDecl[*Database](state, x)
		if err != nil {
			return errors.WithStack(err)
		}
		(*d).Runtime = v
	case *EgressRuntime:
		d, err := findDecl[*Verb](state, x)
		if err != nil {
			return err
		}
		if (*d).Runtime == nil {
			(*d).Runtime = &VerbRuntime{}
		}
		(*d).Runtime.EgressRuntime = v
	}
	return nil
}

func findDecl[T any](module *Module, x *RuntimeElement) (*T, error) {
	name, ok := x.Name.Get()
	if !ok {
		return nil, errors.Errorf("missing name in element")
	}
	for _, decl := range module.Decls {
		if decl.GetName() == name {
			if d, ok := decl.(T); ok {
				return &d, nil
			}
			return nil, errors.Errorf("unexpected decl type %T", decl)
		}
	}
	return nil, errors.Errorf("decl %s not found", name)
}
