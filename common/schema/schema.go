//protobuf:package xyz.block.ftl.schema.v1
//protobuf:option go_package="github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1;schemapb"
//protobuf:option java_multiple_files=true
package schema

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/internal/key"
)

var ErrNotFound = errors.New("not found")

//protobuf:export
type Schema struct {
	Pos    Position `parser:"" protobuf:"1,optional"`
	Realms []*Realm `parser:"@@*" protobuf:"2"`
}

var _ Node = (*Schema)(nil)

func (s *Schema) Position() Position { return s.Pos }
func (s *Schema) String() string {
	out := &strings.Builder{}
	for i, r := range s.Realms {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, r)
	}
	return out.String()
}

func (s *Schema) schemaChildren() []Node {
	var realms []Node
	for _, r := range s.Realms {
		realms = append(realms, r)
	}
	return realms
}

func (s *Schema) Hash() [sha256.Size]byte {
	return sha256.Sum256([]byte(s.String()))
}

// ResolveRequestResponseType resolves a reference to a supported request/response type, which can be a Data or an Any,
// or a TypeAlias over either supported type.
func (s *Schema) ResolveRequestResponseType(ref *Ref) (Symbol, error) {
	decl, ok := s.Resolve(ref).Get()
	if !ok {
		return nil, fmt.Errorf("unknown ref %s", ref)
	}

	if ta, ok := decl.(*TypeAlias); ok {
		if typ, ok := ta.Type.(*Any); ok {
			return typ, nil
		}
	}

	return s.resolveToSymbolMonomorphised(ref, nil)
}

// ResolveMonomorphised resolves a reference to a monomorphised Data type.
// Also supports resolving the monomorphised Data type underlying a TypeAlias, where applicable.
//
// If a Ref is not found, returns ErrNotFound.
func (s *Schema) ResolveMonomorphised(ref *Ref) (*Data, error) {
	return s.resolveToDataMonomorphised(ref, nil)
}

func (s *Schema) resolveToDataMonomorphised(n Node, parent Node) (*Data, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, fmt.Errorf("unknown ref %s", typ)
		}
		return s.resolveToDataMonomorphised(resolved, typ)
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, fmt.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return typ.Monomorphise(p)
	case *TypeAlias:
		return s.resolveToDataMonomorphised(typ.Type, typ)
	default:
		return nil, fmt.Errorf("expected data or type alias of data, got %T", typ)
	}
}

func (s *Schema) resolveToSymbolMonomorphised(n Node, parent Node) (Symbol, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, fmt.Errorf("unknown ref %s", typ)
		}
		return s.resolveToSymbolMonomorphised(resolved, typ)
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, fmt.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return typ.Monomorphise(p)
	case *TypeAlias:
		return s.resolveToSymbolMonomorphised(typ.Type, typ)
	case Symbol:
		return typ, nil
	default:
		return nil, fmt.Errorf("expected data or type alias of data, got %T", typ)
	}
}

// ResolveWithModule a reference to a declaration and its module.
func (s *Schema) ResolveWithModule(ref *Ref) (optional.Option[Decl], optional.Option[*Module]) {
	for _, realm := range s.Realms {
		resolved, module := realm.ResolveWithModule(ref)
		if _, ok := resolved.Get(); ok {
			return resolved, module
		}
	}
	return optional.None[Decl](), optional.None[*Module]()
}

// Resolve a reference to a declaration.
func (s *Schema) Resolve(ref *Ref) optional.Option[Decl] {
	decl, _ := s.ResolveWithModule(ref)
	return decl
}

func (s *Schema) ResolveToType(ref *Ref, out Decl) error {
	for _, realm := range s.Realms {
		if realm.ContainsRef(ref) {
			return realm.ResolveToType(ref, out)
		}
	}
	return fmt.Errorf("could not resolve reference %v: %w", ref, ErrNotFound)
}

func (s *Schema) Module(name string) optional.Option[*Module] {
	for _, realm := range s.Realms {
		module := realm.Module(name)
		if _, ok := module.Get(); ok {
			return module
		}
	}
	return optional.None[*Module]()
}

func (s *Schema) Deployment(name key.Deployment) optional.Option[*Module] {
	for _, realm := range s.Realms {
		deployment := realm.Deployment(name)
		if _, ok := deployment.Get(); ok {
			return deployment
		}
	}
	return optional.None[*Module]()
}

// TypeName returns the name of a type as a string, stripping any package prefix and correctly handling Ref aliases.
func TypeName(v any) string {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()

	// handle AbstractRefs like "AbstractRef[github.com/block/ftl/common/protos/xyz/block/ftl/schema.DataRef]"
	if strings.HasPrefix(t.Name(), "AbstractRef[") {
		return strings.TrimSuffix(strings.Split(t.Name(), ".")[2], "]")
	}

	return t.Name()
}

// FromProto converts a protobuf Schema to a Schema and validates it.
func FromProto(s *schemapb.Schema) (*Schema, error) {
	realms, err := realmListToSchema(s.Realms)
	if err != nil {
		return nil, err
	}
	if len(realms) != 1 {
		return nil, errors.New("expected exactly one realm in schema")
	}
	schema := &Schema{Realms: realms}
	return schema.Validate()
}

func (s *Schema) ModuleDependencies(module string) map[string]*Module {
	for _, realm := range s.Realms {
		deps := realm.ModuleDependencies(module)
		if len(deps) > 0 {
			return deps
		}
	}
	return nil
}

func ValidatedModuleFromProto(v *schemapb.Module) (*Module, error) {
	module, err := ModuleFromProto(v)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal module: %w", err)
	}
	if err := module.Validate(); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *Schema) InternalRealms() []*Realm {
	var out []*Realm
	for _, r := range s.Realms {
		if !r.External {
			out = append(out, r)
		}
	}
	return out
}

func (s *Schema) InternalModules() []*Module {
	var out []*Module
	for _, r := range s.InternalRealms() {
		out = append(out, r.Modules...)
	}
	return out
}

// SchemaState is the schema service state as persisted in Raft
//
//protobuf:export
type SchemaState struct {
	Modules          []*Module                 `protobuf:"1"`
	Changesets       []*Changeset              `protobuf:"2"`
	ChangesetEvents  []*DeploymentRuntimeEvent `protobuf:"3"`
	DeploymentEvents []*DeploymentRuntimeEvent `protobuf:"4"`
	Realms           []*RealmState             `protobuf:"5"`
}

type RealmState struct {
	Name     string `protobuf:"1"`
	External bool   `protobuf:"2"`
}
