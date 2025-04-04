package schema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

type Realm struct {
	Pos Position `parser:"" protobuf:"1"`

	External bool      `parser:"@('external')? 'realm'" protobuf:"2"`
	Name     string    `parser:"@Ident '{'" protobuf:"3"`
	Modules  []*Module `parser:"@@* '}'" protobuf:"4"`
}

var _ Node = (*Realm)(nil)

func (r *Realm) Position() Position { return r.Pos }
func (r *Realm) String() string {
	out := &strings.Builder{}
	if r.External {
		fmt.Fprintf(out, "external realm %s {", r.Name)
	} else {
		fmt.Fprintf(out, "realm %s {", r.Name)
	}
	for _, module := range r.Modules {
		fmt.Fprintln(out)
		fmt.Fprintln(out, indent(module.String()))
	}
	fmt.Fprintf(out, "}\n")
	return out.String()
}

func (r *Realm) schemaChildren() []Node {
	out := make([]Node, len(r.Modules))
	for i, m := range r.Modules {
		out[i] = m
	}
	return out
}

func (r *Realm) ResolveWithModule(ref *Ref) (optional.Option[Decl], optional.Option[*Module]) {
	for _, module := range r.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					return optional.Some(decl), optional.Some(module)
				}
			}
		}
	}
	return optional.None[Decl](), optional.None[*Module]()
}

func (r *Realm) ContainsRef(ref *Ref) bool {
	for _, module := range r.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					return true
				}
			}
		}
	}
	return false
}

// ResolveToType resolves a reference to a declaration of the given type.
//
// The out parameter must be a pointer to a non-nil Decl implementation or this
// will panic.
//
//	data := &Data{}
//	err := s.ResolveToType(ref, data)
func (r *Realm) ResolveToType(ref *Ref, out Decl) error {
	// Programmer error.
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		panic(fmt.Errorf("out parameter is not a pointer"))
	}
	if reflect.ValueOf(out).Elem().Kind() == reflect.Invalid {
		panic(fmt.Errorf("out parameter is a nil pointer"))
	}

	for _, module := range r.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					declType := reflect.TypeOf(decl)
					outType := reflect.TypeOf(out)
					if declType.Elem().AssignableTo(outType.Elem()) {
						reflect.ValueOf(out).Elem().Set(reflect.ValueOf(decl).Elem())
						return nil
					}
					return fmt.Errorf("resolved declaration is not of the expected type: want %s, got %s",
						outType, declType)
				}
			}
		}
	}

	return fmt.Errorf("could not resolve reference %v: %w", ref, ErrNotFound)
}

// Module returns the named module if it exists.
func (r *Realm) Module(name string) optional.Option[*Module] {
	for _, module := range r.Modules {
		if module.Name == name {
			return optional.Some(module)
		}
	}
	return optional.None[*Module]()
}

// Deployment returns the named deployment if it exists.
func (r *Realm) Deployment(name key.Deployment) optional.Option[*Module] {
	for _, module := range r.Modules {
		if module.GetRuntime().GetDeployment().GetDeploymentKey() == name {
			return optional.Some(module)
		}
	}
	return optional.None[*Module]()
}

// Upsert inserts or replaces a module.
func (r *Realm) Upsert(module *Module) {
	for i, m := range r.Modules {
		if m.Name == module.Name {
			r.Modules[i] = module
			return
		}
	}
	r.Modules = append(r.Modules, module)
}

// ModuleDependencies returns the modules that the given module depends on
// Dependency modules are the ones that are called by the given module, or that publish topics that the given module subscribes to
func (r *Realm) ModuleDependencies(module string) map[string]*Module {
	mods := map[string]*Module{}
	for _, sch := range r.Modules {
		mods[sch.Name] = sch
	}
	deps := make(map[string]*Module)
	toProcess := []string{module}
	for len(toProcess) > 0 {
		dep := toProcess[0]
		toProcess = toProcess[1:]
		if deps[dep] != nil {
			continue
		}
		dm := mods[dep]
		deps[dep] = dm
		for _, m := range dm.Decls {
			if ref, ok := m.(*Verb); ok {
				for _, ref := range ref.Metadata {
					switch md := ref.(type) {
					case *MetadataCalls:
						for _, calls := range md.Calls {
							if calls.Module != "" {
								toProcess = append(toProcess, calls.Module)
							}
						}
					case *MetadataSubscriber:
						if md.Topic.Module != "" {
							toProcess = append(toProcess, md.Topic.Module)
						}
					default:
					}
				}
			}
		}
	}
	delete(deps, module)
	return deps
}
