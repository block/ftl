package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// A Data structure.
//
//protobuf:1
type Data struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments       []string         `parser:"@Comment*" protobuf:"2"`
	Export         bool             `parser:"@'export'?" protobuf:"3"`
	Name           string           `parser:"'data' @Ident" protobuf:"4"`
	TypeParameters []*TypeParameter `parser:"( '<' @@ (',' @@)* '>' )?" protobuf:"5"`
	Fields         []*Field         `parser:"'{' @@* '}'" protobuf:"6"`
	Metadata       []Metadata       `parser:"@@*" protobuf:"7"`
}

var _ Decl = (*Data)(nil)
var _ Symbol = (*Data)(nil)
var _ Scoped = (*Data)(nil)

func (d *Data) Scope() Scope {
	scope := Scope{}
	for _, t := range d.TypeParameters {
		scope[t.Name] = ModuleDecl{Symbol: t}
	}
	return scope
}

func (d *Data) FieldByName(name string) *Field {
	for _, f := range d.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// Monomorphise this data type with the given type arguments.
//
// If this data type has no type parameters, it will be returned as-is.
//
// This will return a new Data structure with all type parameters replaced with
// the given types.
func (d *Data) Monomorphise(ref *Ref) (*Data, error) {
	if len(d.TypeParameters) != len(ref.TypeParameters) {
		return nil, fmt.Errorf("%s: expected %d type arguments, got %d", ref.Pos, len(d.TypeParameters), len(ref.TypeParameters))
	}
	if len(d.TypeParameters) == 0 {
		return d, nil
	}
	names := map[string]Type{}
	for i, t := range d.TypeParameters {
		names[t.Name] = ref.TypeParameters[i]
	}
	monomorphised := reflect.DeepCopy(d)
	monomorphised.TypeParameters = nil

	// Because we don't have parent links in the AST allowing us to visit on
	// Type and replace it on the parent, we have to do a full traversal to find
	// the parents of all the Type nodes we need to replace. This will be a bit
	// tricky to maintain, but it's basically any type that has parametric
	// types: maps, slices, fields, etc.
	err := Visit(monomorphised, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *Map:
			k, err := maybeMonomorphiseType(n.Key, names)
			if err != nil {
				return fmt.Errorf("%s: map key: %w", n.Key.Position(), err)
			}
			v, err := maybeMonomorphiseType(n.Value, names)
			if err != nil {
				return fmt.Errorf("%s: map value: %w", n.Value.Position(), err)
			}
			n.Key = k
			n.Value = v

		case *Array:
			t, err := maybeMonomorphiseType(n.Element, names)
			if err != nil {
				return fmt.Errorf("%s: array element: %w", n.Element.Position(), err)
			}
			n.Element = t

		case *Field:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return fmt.Errorf("%s: field type: %w", n.Type.Position(), err)
			}
			n.Type = t

		case *Optional:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return fmt.Errorf("%s: optional type: %w", n.Type.Position(), err)
			}
			n.Type = t

		case *Config:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return fmt.Errorf("%s: config type: %w", n.Type.Position(), err)
			}
			n.Type = t

		case *Secret:
			t, err := maybeMonomorphiseType(n.Type, names)
			if err != nil {
				return fmt.Errorf("%s: secret type: %w", n.Type.Position(), err)
			}
			n.Type = t

		case *Any, *Bool, *Bytes, *Data, *Ref, *Database, Decl, *Float,
			IngressPathComponent, *IngressPathLiteral, *IngressPathParameter,
			*Int, Metadata, *MetadataCalls, *MetadataConfig, *MetadataDatabases, *MetadataRetry,
			*MetadataIngress, *MetadataCronJob, *MetadataAlias, *MetadataSecrets, *Module,
			*Schema, *String, *Time, Type, *TypeParameter, *Unit, *Verb, *Enum,
			*EnumVariant, Value, *IntValue, *StringValue, *TypeValue, Symbol,
			Named, *TypeAlias, *Topic, *MetadataSubscriber, *MetadataTypeMap,
			*MetadataEncoding, *MetadataPublisher, *MetadataSQLMigration, *DSNDatabaseConnector, *DatabaseRuntime,
			DatabaseConnector, *AWSIAMAuthDatabaseConnector, *DatabaseRuntimeConnections, *MetadataArtefact, Provisioned:
		}
		return next()
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to monomorphise: %w", d.Pos, err)
	}
	return monomorphised, nil
}

func (d *Data) Position() Position { return d.Pos }
func (*Data) schemaDecl()          {}
func (*Data) schemaSymbol()        {}
func (d *Data) schemaChildren() []Node {
	children := make([]Node, 0, len(d.Fields)+len(d.Metadata)+len(d.TypeParameters))
	for _, t := range d.TypeParameters {
		children = append(children, t)
	}
	for _, f := range d.Fields {
		children = append(children, f)
	}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}

func (d *Data) GetName() string  { return d.Name }
func (d *Data) IsExported() bool { return d.Export }

func (d *Data) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	typeParameters := ""
	if len(d.TypeParameters) > 0 {
		typeParameters = "<"
		for i, t := range d.TypeParameters {
			if i != 0 {
				typeParameters += ", "
			}
			typeParameters += t.String()
		}
		typeParameters += ">"
	}
	if d.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "data %s%s {\n", d.Name, typeParameters)
	for _, f := range d.Fields {
		fmt.Fprintln(w, indent(f.String()))
	}
	fmt.Fprintf(w, "}")
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

func DataFromProto(s *schemapb.Data) *Data {
	return &Data{
		Pos: PosFromProto(s.Pos),

		Name:           s.Name,
		Export:         s.Export,
		TypeParameters: typeParametersToSchema(s.TypeParameters),
		Fields:         fieldListToSchema(s.Fields),
		Comments:       s.Comments,
		Metadata:       metadataListToSchema(s.Metadata),
	}
}

// MonoType returns the monomorphised type of this data type if applicable, or returns the original type.
func maybeMonomorphiseType(t Type, typeParameters map[string]Type) (Type, error) {
	if t, ok := t.(*Ref); ok && t.Module == "" {
		if tp, ok := typeParameters[t.Name]; ok {
			return tp, nil
		}
		return nil, fmt.Errorf("%s: unknown type parameter %q", t.Position(), t)
	}
	return t, nil
}
