//protobuf:package xyz.block.ftl.schema.v1
//protobuf:option go_package="github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1;schemapb"
//protobuf:option java_multiple_files=true
package schema

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"slices"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/key"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	ftlreflect "github.com/block/ftl/common/reflect"
	ftlslices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/iterops"
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
		return nil, errors.Errorf("unknown ref %s", ref)
	}

	if ta, ok := decl.(*TypeAlias); ok {
		if typ, ok := ta.Type.(*Any); ok {
			return typ, nil
		}
	}

	return errors.WithStack2(s.resolveToSymbolMonomorphised(ref, nil))
}

// ResolveMonomorphised resolves a reference to a monomorphised Data type.
// Also supports resolving the monomorphised Data type underlying a TypeAlias, where applicable.
//
// If a Ref is not found, returns ErrNotFound.
func (s *Schema) ResolveMonomorphised(ref *Ref) (*Data, error) {
	return errors.WithStack2(s.resolveToDataMonomorphised(ref, nil))
}

func (s *Schema) resolveToDataMonomorphised(n Node, parent Node) (*Data, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, errors.Errorf("unknown ref %s", typ)
		}
		return errors.WithStack2(s.resolveToDataMonomorphised(resolved, typ))
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, errors.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return errors.WithStack2(typ.Monomorphise(p))
	case *TypeAlias:
		return errors.WithStack2(s.resolveToDataMonomorphised(typ.Type, typ))
	default:
		return nil, errors.Errorf("expected data or type alias of data, got %T", typ)
	}
}

func (s *Schema) resolveToSymbolMonomorphised(n Node, parent Node) (Symbol, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, errors.Errorf("unknown ref %s", typ)
		}
		return errors.WithStack2(s.resolveToSymbolMonomorphised(resolved, typ))
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, errors.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return errors.WithStack2(typ.Monomorphise(p))
	case *TypeAlias:
		return errors.WithStack2(s.resolveToSymbolMonomorphised(typ.Type, typ))
	case Symbol:
		return typ, nil
	default:
		return nil, errors.Errorf("expected data or type alias of data, got %T", typ)
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

// ResolveType resolves a Ref to a Type, or errors if the dereferenced declaration is not a Type.
//
// If the resolved Type is generic it will be monomorphised.
func (s *Schema) ResolveType(ref *Ref) (Type, error) {
	maybeDecl, _ := s.ResolveWithModule(ref)
	decl, ok := maybeDecl.Get()
	if !ok {
		return nil, errors.Wrapf(ErrNotFound, "could not resolve reference %s", ref)
	}
	dt, ok := decl.(Type)
	if !ok {
		return nil, errors.Errorf("%s: expected type, got %T", ref.Pos, decl)
	}
	if _, ok := dt.(*Data); ok {
		return errors.WithStack2(s.ResolveMonomorphised(ref))
	}
	return dt, nil
}

func (s *Schema) ResolveToType(ref *Ref, out Decl) error {
	for _, realm := range s.Realms {
		if realm.ContainsRef(ref) {
			return errors.WithStack(realm.ResolveToType(ref, out))
		}
	}
	return errors.Wrapf(ErrNotFound, "could not resolve reference %s", ref)
}

func (s *Schema) Module(realm, name string) optional.Option[*Module] {
	for _, r := range s.Realms {
		if r.Name != realm {
			continue
		}
		module := r.Module(name)
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
		return nil, errors.WithStack(err)
	}
	schema := &Schema{Realms: realms}
	return errors.WithStack2(schema.Validate())
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
		return nil, errors.Wrap(err, "failed to unmarshal module")
	}
	if err := module.Validate(); err != nil {
		return nil, errors.WithStack(err)
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

func (s *Schema) FirstInternalRealm() optional.Option[*Realm] {
	for _, r := range s.Realms {
		if !r.External {
			return optional.Some(r)
		}
	}
	return optional.None[*Realm]()
}

func (s *Schema) InternalModules() []*Module {
	var out []*Module
	for _, r := range s.InternalRealms() {
		out = append(out, r.Modules...)
	}
	return out
}

// WithBuiltins returns a new schema with the builtins module added to each internal realm.
func (s *Schema) WithBuiltins() *Schema {
	c := ftlreflect.DeepCopy(s)
	builtins := Builtins()
	for _, r := range c.Realms {
		if !r.External {
			hasBuiltin := false
			for _, m := range r.Modules {
				if m.Name == builtins.Name {
					hasBuiltin = true
				}
			}
			if !hasBuiltin {
				r.Modules = append([]*Module{builtins}, r.Modules...)
			}
		}
	}
	return c
}

// FilterModules returns a new schema with only the modules in the given list of references.
// Returns the filtered schema and any refs that were not found in the original schema.
func (s *Schema) FilterModules(modules []RefKey) (*Schema, []RefKey) {
	filtered := &Schema{}
	moduleNames := make(map[RefKey]bool)
	missing := make(map[RefKey]bool)
	for _, module := range modules {
		moduleNames[module.ModuleOnly()] = true
		missing[module.ModuleOnly()] = true
	}
	for _, realm := range s.Realms {
		filteredRealm := &Realm{
			Name:     realm.Name,
			External: realm.External,
		}
		for _, module := range realm.Modules {
			ref := RefKey{Module: module.Name}
			if moduleNames[ref] {
				filteredRealm.Modules = append(filteredRealm.Modules, module)
				delete(missing, ref)
			}
		}
		if len(filteredRealm.Modules) > 0 {
			filtered.Realms = append(filtered.Realms, filteredRealm)
		}
	}
	return filtered, maps.Keys(missing)
}

// External returns the subset of the schema exported by internal realms.
// This schema can be used as external schema in other clusters.
func (s *Schema) External() *Schema {
	filtered := &Schema{}
	for _, realm := range s.Realms {
		realmCopy := ftlreflect.DeepCopy(realm)
		if !realmCopy.External {
			filtered.Realms = append(
				filtered.Realms,
				realmCopy.FilterByVisibility(VisibilityScopeRealm),
			)
		}
	}
	// strip any runtime, and unsupported decls
	for _, realm := range filtered.Realms {
		for _, module := range realm.Modules {
			module.Runtime = nil
			var filteredDecls []Decl
			for _, decl := range module.Decls {
				switch d := decl.(type) {
				case *Verb:
					d.Runtime = nil
					d.Metadata = filterExternalMetadata(d.Metadata)
					filteredDecls = append(filteredDecls, d)
				case *Database:
					d.Runtime = nil
					d.Metadata = filterExternalMetadata(d.Metadata)
					filteredDecls = append(filteredDecls, d)
				case *Topic:
					d.Runtime = nil
					d.Metadata = filterExternalMetadata(d.Metadata)
					filteredDecls = append(filteredDecls, d)
				case *Data:
					d.Metadata = filterExternalMetadata(d.Metadata)
					filteredDecls = append(filteredDecls, d)
				case *TypeAlias:
					d.Metadata = filterExternalMetadata(d.Metadata)
					filteredDecls = append(filteredDecls, d)
				case *Enum:
					filteredDecls = append(filteredDecls, d)
				case *Config:
				case *Secret:
				}
				module.Decls = filteredDecls
			}
		}
	}
	return filtered
}

func (s *Schema) UpdateRealms(realms []*Realm) {
	byName := make(map[string]*Realm)
	found := make(map[string]bool)
	for _, r := range realms {
		byName[r.Name] = r
		found[r.Name] = true
	}
	for i, r := range s.Realms {
		if n, ok := byName[r.Name]; ok {
			s.Realms[i] = n
		}
		found[r.Name] = true
	}
	for _, r := range realms {
		if !found[r.Name] {
			s.Realms = append(s.Realms, r)
		}
	}
}

func (s *Schema) RemoveModule(realm, module string) {
	for _, r := range s.Realms {
		if r.Name != realm {
			continue
		}
		r.Modules = ftlslices.Filter(r.Modules, func(m *Module) bool {
			return m.Name != module
		})
	}
}

func (s *Schema) Realm(name string) optional.Option[*Realm] {
	for _, r := range s.Realms {
		if r.Name == name {
			return optional.Some(r)
		}
	}
	return optional.None[*Realm]()
}

func filterExternalMetadata(metadata []Metadata) []Metadata {
	var filtered []Metadata
	for _, m := range metadata {
		switch m := m.(type) {
		case *MetadataConfig:
		case *MetadataSecrets:
		default:
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// SchemaState is the schema service state as persisted in Raft
//
//protobuf:export
type SchemaState struct {
	Schema           *Schema                   `protobuf:"1"`
	Changesets       []*Changeset              `protobuf:"2"`
	ChangesetEvents  []*DeploymentRuntimeEvent `protobuf:"3"`
	DeploymentEvents []*DeploymentRuntimeEvent `protobuf:"4"`
}

func (s *SchemaState) Validate() error {
	internals := iterops.Count(slices.Values(s.Schema.Realms), func(r *Realm) bool { return !r.External })
	if internals > 1 {
		return errors.Errorf("only one internal realm is allowed, got %d", internals)
	}
	return nil
}
