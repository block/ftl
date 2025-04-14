package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"

	"github.com/block/ftl/common/strcase"
)

const help = `Generate a Protobuf schema from Go types.

It supports converting structs to messages, Go "sum types" to oneof fields, and Go "enums" to Protobuf enums.

The generator works by scanning a package for types marked with //protobuf:export directives. For each exported type,
it extracts protobuf tags from the source Go types. There are two locations where these tags must be specified:

  1. For fields using a tag in the form ` + "`protobuf:\"<id>[,optional]\"`" + `.
  2. For sum types as comment directives in the form //protobuf:<id>.

An example showing all three supported types and the corresponding protobuf tags:

	//protobuf:export
	type UserType int

	const (
		UserTypeUnknown UserType = iota
		UserTypeAdmin
		UserTypeUser
	)

	// Entity is a "sum type" consisting of User and Group.
	//
	// Every sum type element must have a comment directive in the form //protobuf:<id>.
	//protobuf:export
	type Entity interface { entity() }

	//protobuf:1
	type User struct {
		Name   string    ` + "`protobuf:\"1\"`" + `
		Type   UserType ` + "`protobuf:\"2\"`" + `
	}
	func (User) entity() {}

	//protobuf:2
	type Group struct {
		Users []string ` + "`protobuf:\"1\"`" + `
	}
	func (Group) entity() {}

	//protobuf:export
	type Role struct {
		Name string ` + "`protobuf:\"1\"`" + `
		Entities []Entity ` + "`protobuf:\"2\"`" + `
	}

And this is the corresponding protobuf schema:

	message Entity {
	  oneof value {
	    User user = 1;
	    Group group = 2;
	  }
	}

	enum UserType {
	  USER_TYPE_UNKNOWN = 0;
	  USER_TYPE_ADMIN = 1;
	  USER_TYPE_USER = 2;
	}

	message User {
	  string Name = 1;
	  UserType Type = 2;
	}

	message Group {
	  repeated string users = 1;
	}

	message Role {
	  string name = 1;
	  repeated Entity entities = 2;
	}
`

var (
	// interface { Validate() error }
	validatorIface = types.NewInterfaceType([]*types.Func{ //nolint:forcetypeassert
		types.NewFunc(token.Pos(0), nil, "Validate",
			types.NewSignatureType(nil, nil, nil,
				types.NewTuple(),
				types.NewTuple(types.NewVar(0, nil, "err", types.Universe.Lookup("error").(*types.TypeName).Type())),
				false)),
	}, nil)
	textMarshaler   = loadInterface("encoding", "TextMarshaler")
	binaryMarshaler = loadInterface("encoding", "BinaryMarshaler")
	// stdTypes is a map of Go types to corresponding protobuf types.
	stdTypes = map[string]stdType{
		"time.Time":     {"google.protobuf.Timestamp", "google/protobuf/timestamp.proto"},
		"time.Duration": {"google.protobuf.Duration", "google/protobuf/duration.proto"},
	}
	builtinTypes = map[string]struct{}{
		"bool": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {},
		"uint32": {}, "uint64": {}, "float32": {}, "float64": {}, "string": {},
	}
)

type stdType struct {
	ref  string
	path string
}

type File struct {
	GoPackage string
	Imports   []string
	Decls     []Decl
}

func (f *File) AddImport(name string) {
	if slices.Contains(f.Imports, name) {
		return
	}
	f.Imports = append(f.Imports, name)
}

func (f File) OrderedDecls() []Decl {
	decls := make([]Decl, len(f.Decls))
	copy(decls, f.Decls)
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].DeclName() < decls[j].DeclName()
	})
	return decls
}

// KindOf looks up the kind of a type in the declarations. Returns KindUnspecified if the type is not found.
func (f File) KindOf(t types.Type, name string) Kind {
	pathComponents := strings.Split(t.String(), "/")
	typeComponents := strings.Split(pathComponents[len(pathComponents)-1], ".")
	if len(typeComponents) == 2 && typeComponents[0] == f.GoPackage {
		for _, decl := range f.Decls {
			if decl.DeclName() == name {
				return Kind(reflect.Indirect(reflect.ValueOf(decl)).Type().Name())
			}
		}
	}
	if _, ok := builtinTypes[name]; ok {
		return KindBuiltin
	}
	if _, ok := stdTypes[name]; ok {
		return KindStdlib
	}
	return KindUnspecified
}

//sumtype:decl
type Decl interface {
	decl()
	DeclName() string
}

type Message struct {
	Comment   string
	Name      string
	Validator bool
	Fields    []*Field
}

func (Message) decl()              {}
func (m Message) DeclName() string { return m.Name }

type Kind string

func (k Kind) String() string {
	if k == KindUnspecified {
		return "Unspecified"
	}
	return string(k)
}

const (
	KindUnspecified     Kind = ""
	KindBuiltin         Kind = "Builtin"
	KindStdlib          Kind = "Stdlib"
	KindMessage         Kind = "Message"
	KindEnum            Kind = "Enum"
	KindSumType         Kind = "SumType"
	KindBinaryMarshaler Kind = "BinaryMarshaler"
	KindTextMarshaler   Kind = "TextMarshaler"
	KindMap             Kind = "Map"
)

type Field struct {
	ID          int
	Name        string
	OriginType  string // The original type of the field, eg. int, string, float32, etc.
	ProtoType   string // The type of the field in the generated .proto file.
	ProtoGoType string // The type of the field in the generated Go protobuf code. eg. int -> int64.

	Optional        bool
	OptionalWrapper bool // optional as alecthomas/types/optional.Option
	Repeated        bool
	Pointer         bool

	// Map-specific fields
	MapKey   *Field // Type information for map keys
	MapValue *Field // Type information for map values

	Import    string // required import to create this type
	Converter *TypeConverter
	Kind      Kind
}

func (f Field) OriginTypeSignature() string {
	if f.Pointer {
		return "*" + f.OriginType
	}
	return f.OriginType
}

type TypeConverter struct {
	FromProto func(variable string) string
	ToProto   func(variable string) string

	// ProtoPointer is true if the proto type is a pointer.
	ProtoPointer bool
	// ToProtoTakesPointer is true if the ToProto function takes a pointer as an argument.
	ToProtoTakesPointer bool
}

var reservedWords = map[string]string{
	"String": "String_",
}

func protoName(s string) string {
	if name, ok := reservedWords[s]; ok {
		return name
	}
	return strcase.ToUpperCamel(s)
}

func (f Field) EscapedName() string {
	if name, ok := reservedWords[f.Name]; ok {
		return name
	}
	return strcase.ToUpperCamel(f.Name)
}

type Enum struct {
	Comment string
	Name    string
	Values  map[string]int
}

func (e Enum) ByValue() map[int]string {
	m := map[int]string{}
	for k, v := range e.Values {
		m[v] = k
	}
	return m
}

func (Enum) decl()              {}
func (e Enum) DeclName() string { return e.Name }

type SumType struct {
	Comment  string
	Name     string
	Variants map[string]int
}

func (SumType) decl()              {}
func (s SumType) DeclName() string { return s.Name }

func (s SumType) ShouldStripPrefix() bool {
	for name := range s.Variants {
		if !strings.HasPrefix(name, s.Name) {
			return false
		}
	}
	return true
}

func (s SumType) FieldName(variant string) string {
	if s.ShouldStripPrefix() {
		return strcase.ToLowerSnake(strings.TrimPrefix(variant, s.Name))
	}
	return strcase.ToLowerSnake(variant)
}

func (s SumType) Getter(variant string) string {
	if s.ShouldStripPrefix() {
		return "Get" + protoName(strings.TrimPrefix(variant, s.Name))
	}
	return "Get" + protoName(variant)
}

// TypeName returns the full oneof typename, which is <SumType>_<VariantType>
func (s SumType) TypeName(variant string) string {
	out := ""
	for name := range s.Variants {
		if !strings.HasPrefix(name, s.Name) {
			out = protoName(s.Name) + "_" + protoName(variant)
			break
		}
	}
	if out == "" {
		out = protoName(s.Name) + "_" + protoName(strings.TrimPrefix(variant, s.Name))
	}
	return out
}

// TextMarshaler is a named type that implements encoding.TextMarshaler. Encoding will delegate to the marshaller.
type TextMarshaler struct {
	Comment string
	Name    string
}

func (TextMarshaler) decl()              {}
func (u TextMarshaler) DeclName() string { return u.Name }

// BinaryMarshaler is a named type that implements encoding.BinaryMarshaler. Encoding will delegate to the marshaller.
type BinaryMarshaler struct {
	Comment string
	Name    string
}

func (BinaryMarshaler) decl()              {}
func (u BinaryMarshaler) DeclName() string { return u.Name }

type Config struct {
	Output  string `help:"Output file to write generated protobuf schema to." short:"o" xor:"output"`
	JSON    bool   `help:"Dump intermediate JSON represesentation." short:"j" xor:"output"`
	Mappers bool   `help:"Generate ToProto and FromProto mappers for each message." short:"m"`

	Package string `arg:"" help:"Package to scan for types with //protobuf:export directives" required:"true" placeholder:"PKG"`
}

func main() {
	cli := Config{}
	kctx := kong.Parse(&cli, kong.Description(help), kong.UsageOnError())
	err := run(cli)
	kctx.FatalIfErrorf(err)
}

// findExportedTypes scans a package for types marked with //protobuf:export directives
func findExportedTypes(pkg *packages.Package) []string {
	var exports []string
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if genDecl, ok := n.(*ast.GenDecl); ok {
				if genDecl.Doc == nil {
					return true
				}
				for _, comment := range genDecl.Doc.List {
					if strings.TrimSpace(comment.Text) == "//protobuf:export" {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*ast.TypeSpec); ok {
								exports = append(exports, typeSpec.Name.Name)
							}
						}
					}
				}
			}
			return true
		})
	}
	return exports
}

func run(cli Config) error {
	out := os.Stdout
	if cli.Output != "" {
		var err error
		out, err = os.Create(cli.Output + "~")
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer out.Close()
		defer os.Remove(cli.Output + "~")
	}

	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, cli.Package)
	if err != nil {
		return fmt.Errorf("unable to load package %s: %w", cli.Package, err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found matching %s", cli.Package)
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "go2proto: warning: %s\n", pkg.Errors[0])
	}

	if pkg.Types == nil {
		return fmt.Errorf("package %s had fatal errors, cannot continue", cli.Package)
	}

	exports := findExportedTypes(pkg)
	if len(exports) == 0 {
		return fmt.Errorf("no types found with //protobuf:export directive in package %s", cli.Package)
	}

	resolved := &PkgRefs{
		Path: cli.Package,
		Pkg:  pkg,
		Refs: exports,
	}

	directives, err := parsePackageDirectives(pkgs)
	if err != nil {
		return err
	}

	file, goImports, err := extract(cli, resolved)
	if gerr := new(GenError); errors.As(err, &gerr) {
		pos := fset.Position(gerr.pos)
		return fmt.Errorf("%s:%d: %w", pos.Filename, pos.Line, err)
	} else if err != nil {
		return err
	}

	if cli.JSON {
		b, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	err = render(out, directives, file)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if cli.Mappers {
		w, err := os.CreateTemp(cli.Package, "go2proto.to.go-*")
		if err != nil {
			return fmt.Errorf("create temp: %w", err)
		}
		defer os.Remove(w.Name())
		defer w.Close()
		err = renderToProto(w, directives, file, goImports)
		if err != nil {
			return err
		}
		err = os.Rename(w.Name(), filepath.Join(cli.Package, "go2proto.to.go"))
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	}

	if cli.Output != "" {
		err = os.Rename(cli.Output+"~", cli.Output)
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	}
	return nil
}

type GenError struct {
	pos token.Pos
	err error
}

func (g GenError) Error() string { return g.err.Error() }
func (g GenError) Unwrap() error { return g.err }

type PkgRefs struct {
	Path string
	Ref  string
	Refs []string
	Pkg  *packages.Package
}

type State struct {
	Pass      int
	Messages  map[*Message]*types.Named
	GoImports []string
	Dest      File
	Seen      map[string]bool
	Config
	*PkgRefs
}

func genErrorf(pos token.Pos, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if gerr := new(GenError); errors.As(err, &gerr) {
		return &GenError{pos: gerr.pos, err: err}
	}
	return &GenError{pos: pos, err: err}
}

func extract(config Config, pkg *PkgRefs) (File, []string, error) {
	state := State{
		Messages: map[*Message]*types.Named{},
		Seen:     map[string]bool{},
		Config:   config,
		PkgRefs:  pkg,
	}
	// First pass, extract all the decls.
	for _, sym := range pkg.Refs {
		obj := pkg.Pkg.Types.Scope().Lookup(sym)
		if obj == nil {
			return File{}, nil, fmt.Errorf("%s: not found in package %s", sym, pkg.Pkg.ID)
		}
		if !strings.HasSuffix(pkg.Pkg.Name, "_test") {
			state.Dest.GoPackage = pkg.Pkg.Name
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return File{}, nil, genErrorf(obj.Pos(), "%s: expected named type, got %T", sym, obj.Type())
		}
		if err := state.extractDecl(obj, named); err != nil {
			return File{}, nil, fmt.Errorf("%s: %w", sym, err)
		}
	}
	state.Pass++
	// Second pass, populate the fields of messages.
	for msg, n := range state.Messages {
		if err := state.populateFields(msg, n); err != nil {
			return File{}, nil, fmt.Errorf("%s: %w", msg.Name, err)
		}
	}

	imports := map[string]bool{}
	for msg := range state.Messages {
		for _, field := range msg.Fields {
			if field.MapValue != nil && field.MapValue.Import != "" {
				imports[field.MapValue.Import] = true
			}
		}
	}
	for imp := range imports {
		state.GoImports = append(state.GoImports, imp)
	}

	return state.Dest, state.GoImports, nil
}

func isOptional(named *types.Named) bool {
	path := named.Origin().Obj().Pkg().Path()
	name := named.Origin().Obj().Name()
	return path == "github.com/alecthomas/types/optional" && name == "Option"
}

func (s *State) extractDecl(obj types.Object, named *types.Named) error {
	if isOptional(named) {
		u := named.TypeArgs().At(0)
		if nt, ok := u.(*types.Named); ok {
			return s.extractDecl(nt.Obj(), nt)
		}
		return nil
	}

	if named.TypeParams() != nil {
		return genErrorf(obj.Pos(), "generic types are not supported")
	}
	if imp, ok := stdTypes[named.String()]; ok {
		s.Dest.AddImport(imp.path)
		return nil
	}
	switch u := named.Underlying().(type) {
	case *types.Struct:
		if err := s.extractStruct(named); err != nil {
			return genErrorf(obj.Pos(), "%w", err)
		}
		return nil

	case *types.Interface:
		return s.extractSumType(named.Obj(), u)

	case *types.Basic:
		return s.extractEnum(named)

	default:
		if implements(named, binaryMarshaler) || implements(named, textMarshaler) {
			return nil
		}
		return genErrorf(obj.Pos(), "unsupported named type %T", u)
	}
}

func (s *State) extractStruct(n *types.Named) error {
	name := n.Obj().Name()
	if _, ok := s.Seen[name]; ok {
		return nil
	}
	s.Seen[name] = true
	if implements(n, binaryMarshaler) || implements(n, textMarshaler) {
		return nil
	}
	decl := &Message{
		Name:      name,
		Validator: implements(n, validatorIface),
	}
	if comment := findCommentsForObject(n.Obj().Pos(), s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
	}
	// First pass over structs we just want to extract type information. The fields themselves will be populated in the
	// second pass.
	fields, errf := iterFields(n)
	for rf := range fields {
		if err := s.maybeExtractDecl(rf, rf.Type()); err != nil {
			return err
		}
	}
	if err := errf(); err != nil {
		return err
	}
	s.Messages[decl] = n
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) maybeExtractDecl(n types.Object, t types.Type) error {
	switch t := t.(type) {
	case *types.Named:
		return s.extractDecl(n, t)

	case *types.Array:
		return s.maybeExtractDecl(n, t.Elem())

	case *types.Slice:
		return s.maybeExtractDecl(n, t.Elem())

	case *types.Pointer:
		return s.maybeExtractDecl(n, t.Elem())

	case *types.Interface:
		return s.extractSumType(n, t)

	case *types.Map:
		if err := s.maybeExtractDecl(n, t.Elem()); err != nil {
			return err
		}
		return nil

	default:
		return nil
	}
}

func (s *State) populateFields(decl *Message, n *types.Named) error {
	fields, errf := iterFields(n)
	for rf, tag := range fields {
		field := &Field{
			Name: rf.Name(),
		}
		t := rf.Type()
		if nt, ok := t.(*types.Named); ok && isOptional(nt) {
			field.Optional = true
			field.OptionalWrapper = true
			t = nt.TypeArgs().At(0)
		}
		if err := s.applyFieldType(t, field); err != nil {
			return fmt.Errorf("%s: %w", rf.Name(), err)
		}
		field.ID = tag.ID
		field.Optional = tag.Optional || field.Optional
		if field.Optional && field.Repeated {
			return genErrorf(n.Obj().Pos(), "%s: repeated optional fields are not supported", rf.Name())
		}
		if nt, ok := t.(*types.Named); ok {
			if err := s.extractDecl(rf, nt); err != nil {
				return fmt.Errorf("%s: %w", rf.Name(), err)
			}
		}
		if field.Kind == KindUnspecified {
			field.Kind = s.Dest.KindOf(rf.Type(), field.OriginType)
		}

		s.populateConverters(field)
		decl.Fields = append(decl.Fields, field)
	}
	return errf()
}

func (f *Field) ToProto() string {
	name := "x"
	if fn := f.Name; fn != "" {
		name = fmt.Sprintf("%s.%s", name, fn)
	}
	if f.Optional {
		if f.OptionalWrapper {
			if f.Converter.ToProtoTakesPointer {
				if f.Pointer {
					return f.Converter.ToProto(name + ".Default(nil)")
				}
				return f.Converter.ToProto(name + ".Ptr()")
			}
			return "setNil(" + f.Converter.ToProto("orZero("+name+".Ptr())") + ", " + name + ".Ptr())"
		} else if f.Pointer && !f.Converter.ProtoPointer {
			return "setNil(" + f.Converter.ToProto("orZero("+name+")") + ", " + name + ")"
		}
		return f.Converter.ToProto(name)
	} else if f.Repeated {
		if f.Converter.ProtoPointer {
			return "sliceMap(" + name + ", func(v " + f.OriginTypeSignature() + ") " + f.ProtoGoType + " { return " + f.Converter.ToProto("v") + " })"
		}
		return "sliceMap(" + name + ", func(v " + f.OriginType + ") " + f.ProtoGoType + " { return orZero(" + f.Converter.ToProto("v") + ") })"
	}

	if f.Converter.ProtoPointer || f.Kind == KindMap {
		return f.Converter.ToProto(name)
	}

	return "orZero(" + f.Converter.ToProto(name) + ")"
}

func (f *Field) FromProto() string {
	name := "v"
	if fn := f.EscapedName(); fn != "" {
		name = fmt.Sprintf("%s.%s", name, fn)
	}
	// inputs are result.Result[*T]
	input := f.Converter.FromProto(name)
	if f.Optional {
		if f.OptionalWrapper {
			if f.Pointer {
				return "optionalRPtr(" + input + ")"
			}
			return "optionalR(" + input + ")"
		}
		if !f.Pointer {
			return "orZeroR(" + input + ")"
		}
		return input
	} else if f.Repeated {
		if !f.Pointer {
			return "sliceMapR(" + name + ", func(v " + f.ProtoGoType + ") result.Result[" + f.OriginType + "] { return orZeroR(" + f.Converter.FromProto("v") + ") })"
		}
		return "sliceMapR(" + name + ", func(v " + f.ProtoGoType + ") result.Result[*" + f.OriginType + "] { return " + f.Converter.FromProto("v") + " })"
	} else if f.Kind == KindMap {
		return input
	} else if !f.Pointer {
		return "orZeroR(" + input + ")"
	}
	return input
}

func (s *State) populateConverters(field *Field) {
	if field.Kind == KindMap {
		s.populateConverters(field.MapKey)
		s.populateConverters(field.MapValue)
	}
	s.populateNestedConverters(field)
}

func (s *State) populateNestedConverters(field *Field) {
	if field.ProtoType == "google.protobuf.Timestamp" {
		field.Converter = &TypeConverter{
			FromProto:    func(v string) string { return fmt.Sprintf("result.From(setNil(ptr(%s.AsTime()), %s), nil)", v, v) },
			ToProto:      func(v string) string { return fmt.Sprintf("timestamppb.New(%s)", v) },
			ProtoPointer: true,
		}
		field.ProtoGoType = "*timestamppb.Timestamp"
		field.OriginType = "time.Time"
	} else if field.ProtoType == "google.protobuf.Duration" {
		field.Converter = &TypeConverter{
			FromProto:    func(v string) string { return fmt.Sprintf("result.From(setNil(ptr(%s.AsDuration()), %s), nil)", v, v) },
			ToProto:      func(v string) string { return fmt.Sprintf("durationpb.New(%s)", v) },
			ProtoPointer: true,
		}
		field.ProtoGoType = "*durationpb.Duration"
		field.OriginType = "time.Duration"
	} else if field.Kind == KindMap {
		field.Converter = &TypeConverter{
			FromProto: func(v string) string {
				return fmt.Sprintf("mapValuesR(%s, func(v %s) result.Result[%s] { return %s })",
					v,
					field.MapValue.ProtoGoType,
					field.MapValue.OriginType,
					field.MapValue.FromProto())
			},
			ToProto: func(v string) string {
				return fmt.Sprintf("mapValues(%s, func(x %s) %s { return %s })",
					v,
					field.MapValue.OriginType,
					field.MapValue.ProtoGoType,
					field.MapValue.ToProto())
			},
		}
	} else if field.Kind == KindMessage {
		field.Converter = &TypeConverter{
			FromProto:           func(v string) string { return fmt.Sprintf("result.From(%sFromProto(%s))", field.OriginType, v) },
			ToProto:             func(v string) string { return fmt.Sprintf("%s.ToProto()", v) },
			ProtoPointer:        true,
			ToProtoTakesPointer: true,
		}
	} else if field.Kind == KindEnum {
		field.Converter = &TypeConverter{
			FromProto: func(v string) string { return fmt.Sprintf("ptrR(result.From(%sFromProto(%s)))", field.OriginType, v) },
			ToProto:   func(v string) string { return fmt.Sprintf("ptr(%s.ToProto())", v) },
		}
	} else if field.Kind == KindTextMarshaler {
		field.Converter = &TypeConverter{
			FromProto: func(v string) string {
				if field.Pointer {
					return fmt.Sprintf("unmarshallText([]byte(%s), out.%s)", v, field.Name)
				}
				return fmt.Sprintf("unmarshallText([]byte(%s), &out.%s)", v, field.Name)
			},
			ToProto: func(v string) string { return fmt.Sprintf("ptr(string(protoMust(%s.MarshalText())))", v) },
		}
	} else if field.Kind == KindBinaryMarshaler {
		field.Converter = &TypeConverter{
			FromProto: func(v string) string {
				if field.Pointer {
					return fmt.Sprintf("unmarshallBinary(%s, out.%s)", v, field.Name)
				}
				return fmt.Sprintf("unmarshallBinary(%s, &out.%s)", v, field.Name)
			},
			ToProto: func(v string) string { return fmt.Sprintf("ptr(protoMust(%s.MarshalBinary()))", v) },
		}
	} else if field.Kind == KindSumType {
		field.Converter = &TypeConverter{
			FromProto:           func(v string) string { return fmt.Sprintf("ptrR(result.From(%sFromProto(%s)))", field.OriginType, v) },
			ToProto:             func(v string) string { return fmt.Sprintf("%sToProto(%s)", field.OriginType, v) },
			ProtoPointer:        true,
			ToProtoTakesPointer: true,
		}
	} else {
		if field.Pointer || field.Optional {
			field.Converter = &TypeConverter{
				FromProto: func(v string) string {
					return fmt.Sprintf("result.From(setNil(ptr(%s(orZero(%s))), %s), nil)", field.OriginType, v, v)
				},
				ToProto: func(v string) string { return fmt.Sprintf("ptr(%s(%s))", field.ProtoGoType, v) },
			}
		} else {
			field.Converter = &TypeConverter{
				FromProto: func(v string) string { return fmt.Sprintf("result.From(ptr(%s(%s)), nil)", field.OriginType, v) },
				ToProto:   func(v string) string { return fmt.Sprintf("ptr(%s(%s))", field.ProtoGoType, v) },
			}
		}
	}
}

func (s *State) extractSumType(obj types.Object, i *types.Interface) error {
	sumTypeName := obj.Name()
	if _, ok := s.Seen[sumTypeName]; ok {
		return nil
	}
	s.Seen[sumTypeName] = true
	decl := SumType{
		Name:     sumTypeName,
		Variants: map[string]int{},
	}
	if comment := findCommentsForObject(obj.Pos(), s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
	}
	scope := s.Pkg.Types.Scope()
	for _, name := range scope.Names() {
		sym := scope.Lookup(name)
		if sym == obj || (!types.Implements(sym.Type(), i) && !types.Implements(types.NewPointer(sym.Type()), i)) {
			continue
		}

		var pbDirectives []*pbTag
		interfaceType := false
		if _, ok := sym.Type().Underlying().(*types.Interface); ok {
			interfaceType = true
		}

		if comments := findCommentsForObject(sym.Pos(), s.Pkg.Syntax); comments != nil {
			for _, line := range comments.List {
				if strings.HasPrefix(line.Text, "//protobuf:") {
					tag, ok, err := parsePBTag(strings.TrimPrefix(line.Text, "//protobuf:"))
					if err != nil {
						return genErrorf(sym.Pos(), "invalid //protobuf: directive %q: %w", line.Text, err)
					}
					if !ok {
						continue
					}
					pbDirectives = append(pbDirectives, &tag)
				}
			}
		}
		if len(pbDirectives) == 0 {
			// skip interface types. These would result into nested oneofs. We only include leafs as a flat list.
			if !interfaceType {
				return genErrorf(sym.Pos(), "sum type element is missing //protobuf:<id> directive: %s", sym.Name())
			}
		}
		if err := s.extractDecl(sym, sym.Type().(*types.Named)); err != nil { //nolint:forcetypeassert
			return genErrorf(sym.Pos(), "%s: %w", name, err)
		}
		id := -1
		for _, directive := range pbDirectives {
			if id < 0 && directive.SumType == "" {
				id = directive.ID
			} else if directive.SumType == sumTypeName {
				id = directive.ID
			}
		}
		if !interfaceType {
			// we do not want to repeat both sumtypes and their elements
			decl.Variants[name] = id
		}
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) extractEnum(t *types.Named) error {
	if imp, ok := stdTypes[t.String()]; ok {
		s.Dest.AddImport(imp.path)
		return nil
	}
	enumName := t.Obj().Name()
	if _, ok := s.Seen[enumName]; ok {
		return nil
	}
	s.Seen[enumName] = true
	decl := Enum{
		Name:   enumName,
		Values: map[string]int{},
	}
	if comment := findCommentsForObject(t.Obj().Pos(), s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
	}
	scope := s.Pkg.Types.Scope()
	for _, name := range scope.Names() {
		sym := scope.Lookup(name)
		if sym == t.Obj() || sym.Type() != t {
			continue
		}
		c, ok := sym.(*types.Const)
		if !ok {
			return genErrorf(sym.Pos(), "expected const")
		}
		n, err := strconv.Atoi(c.Val().String())
		if err != nil {
			return genErrorf(sym.Pos(), "enum value %q must be a constant integer: %w", c.Val(), err)
		}
		if !strings.HasPrefix(name, enumName) {
			return genErrorf(sym.Pos(), "enum value %q must start with %q", name, enumName)
		}
		decl.Values[name] = n
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) canMarshal(t types.Type, field *Field, name string) bool {
	if _, ok := stdTypes[name]; !ok && s.Dest.KindOf(t, name) == KindUnspecified {
		if implements(t, textMarshaler) {
			field.ProtoType = "string"
			field.ProtoGoType = "string"
			field.Kind = KindTextMarshaler
			return true
		} else if implements(t, binaryMarshaler) {
			field.ProtoType = "bytes"
			field.ProtoGoType = "bytes"
			field.Kind = KindBinaryMarshaler
			return true
		}
	}
	return false
}

func originType(t types.Type) string {
	str := t.String()
	if strings.Contains(str, "/") {
		parts := strings.Split(str, "/")
		return parts[len(parts)-1]
	}
	return str
}

func importStr(t types.Type) string {
	str := t.String()
	if strings.Contains(str, "/") {
		parts := strings.Split(str, ".")
		return strings.TrimPrefix(strings.Join(parts[:len(parts)-1], "."), "*")
	}
	if _, ok := stdTypes[str]; ok {
		parts := strings.Split(str, ".")
		return parts[0]
	}
	return ""
}

func (s *State) applyFieldType(t types.Type, field *Field) error {
	field.OriginType = originType(t)
	field.Import = importStr(t)

	// remove imports to the current package
	if field.Import == s.PkgRefs.Pkg.String() {
		field.Import = ""
	}

	switch t := t.(type) {
	case *types.Alias:
		if s.canMarshal(t, field, t.Obj().Name()) {
			return nil
		}

	case *types.Named:
		ref := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		if bt, ok := stdTypes[ref]; ok {
			field.ProtoType = bt.ref
			field.ProtoGoType = protoName(bt.ref)
			field.OriginType = t.Obj().Name()
			field.Kind = KindStdlib
		} else {
			if s.canMarshal(t, field, t.Obj().Name()) {
				return nil
			}
			if err := s.extractDecl(t.Obj(), t); err != nil {
				return err
			}
			field.ProtoType = t.Obj().Name()
			field.ProtoGoType = "*destpb." + protoName(t.Obj().Name())
			field.OriginType = t.Obj().Name()
		}

	case *types.Array:
		if t.Elem().String() == "byte" {
			field.ProtoType = "bytes"
		} else {
			field.Repeated = true
			return s.applyFieldType(t.Elem(), field)
		}

	case *types.Slice:
		if t.Elem().String() == "byte" {
			field.ProtoType = "bytes"
		} else {
			field.Repeated = true
			return s.applyFieldType(t.Elem(), field)
		}

	case *types.Pointer:
		field.Pointer = true
		if _, ok := t.Elem().(*types.Slice); ok {
			return fmt.Errorf("pointer to slice is not supported")
		}
		return s.applyFieldType(t.Elem(), field)

	case *types.Basic:
		field.ProtoType = t.String()
		field.ProtoGoType = t.String()
		field.Kind = KindBuiltin
		switch t.String() {
		case "int":
			field.ProtoType = "int64"
			field.ProtoGoType = "int64"

		case "uint":
			field.ProtoType = "uint64"
			field.ProtoGoType = "uint64"

		case "float64":
			field.ProtoType = "double"

		case "float32":
			field.ProtoType = "float"

		case "string", "bool", "uint64", "int64", "uint32", "int32":

		default:
			return fmt.Errorf("unsupported basic type %s", t)

		}
	case *types.Map:
		keyType := t.Key()
		valueType := t.Elem()

		field.MapKey = &Field{}
		field.MapValue = &Field{}

		if err := s.applyFieldType(keyType, field.MapKey); err != nil {
			return fmt.Errorf("invalid map key type: %w", err)
		}
		if err := s.applyFieldType(valueType, field.MapValue); err != nil {
			return fmt.Errorf("invalid map value type: %w", err)
		}

		// only integral or string types allowed as map keys (https://protobuf.dev/programming-guides/proto3/)
		switch field.MapKey.ProtoType {
		case "int32", "int64", "uint32", "uint64", "sint32", "sint64",
			"fixed32", "fixed64", "sfixed32", "sfixed64", "string":
		default:
			return fmt.Errorf("invalid map key type %s: map keys must be integral or string types", field.MapKey.ProtoType)
		}

		field.ProtoType = fmt.Sprintf("map<%s, %s>", field.MapKey.ProtoType, field.MapValue.ProtoType)
		field.ProtoGoType = fmt.Sprintf("map[%s]%s", field.MapKey.ProtoGoType, field.MapValue.ProtoGoType)
		field.OriginType = fmt.Sprintf("map[%s]%s", field.MapKey.OriginType, field.MapValue.OriginType)
		field.Kind = KindMap
		field.Import = field.MapValue.Import

	default:
		return fmt.Errorf("unsupported type %s (%T)", t, t)
	}
	return nil
}

func findCommentsForObject(pos token.Pos, syntax []*ast.File) *ast.CommentGroup {
	for _, file := range syntax {
		if file.Pos() <= pos && pos <= file.End() {
			// Use ast.Inspect to traverse the AST and locate the node
			var comments *ast.CommentGroup
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				// If found, get the documentation comments
				if node, ok := n.(*ast.GenDecl); ok {
					for _, spec := range node.Specs {
						if spec.Pos() == pos {
							comments = node.Doc
							return false // Stop the traversal once the node is found
						}
					}
				}
				return true
			})
			return comments
		}
	}
	return nil
}

func implements(t types.Type, i *types.Interface) bool {
	return types.Implements(t, i) || types.Implements(types.NewPointer(t), i)
}

// Returns an iterator over the fields of a struct, and a function that reports any error that occurred.
func iterFields(n *types.Named) (iter.Seq2[*types.Var, pbTag], func() error) {
	var err error
	return func(yield func(*types.Var, pbTag) bool) {
		t, ok := n.Underlying().(*types.Struct)
		if !ok {
			err = fmt.Errorf("expected struct, got %T", n.Underlying())
			return
		}
		for i := range t.NumFields() {
			rf := t.Field(i)
			pb := reflect.StructTag(t.Tag(i)).Get("protobuf")
			if pb == "-" {
				continue
			} else if pb == "" {
				err = genErrorf(n.Obj().Pos(), "%s: missing protobuf tag", rf.Name())
				return
			}
			var pbt pbTag
			pbt, ok, err = parsePBTag(pb)
			if !ok {
				err = genErrorf(rf.Pos(), "invalid protobuf tag")
			}
			if err != nil {
				return
			}
			if !yield(rf, pbt) {
				return
			}
		}
	}, func() error { return err }
}

type pbTag struct {
	ID       int
	SumType  string
	Optional bool
}

func parsePBTag(tag string) (pbTag, bool, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return pbTag{}, false, fmt.Errorf("missing tag")
	}

	idParts := strings.Split(parts[0], " ")
	sumType := ""
	if len(idParts) > 1 {
		sumType = idParts[1]
	}

	id, err := strconv.Atoi(idParts[0])
	if err != nil {
		return pbTag{}, false, nil
	}
	out := pbTag{ID: id, SumType: sumType}
	for _, part := range parts[1:] {
		switch part {
		case "optional":
			out.Optional = true

		default:
			return pbTag{}, false, fmt.Errorf("unknown tag: %s", tag)
		}
	}
	return out, true, nil
}

func loadInterface(pkg, symbol string) *types.Interface {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, pkg)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		for _, name := range pkg.TypesInfo.Defs {
			if t, ok := name.(*types.TypeName); ok {
				if t.Name() == symbol {
					return t.Type().Underlying().(*types.Interface) //nolint:forcetypeassert
				}
			}
		}
	}
	panic("could not find " + pkg + "." + symbol)
}

// PackageDirectives captures the directives in the protobuf:XYZ directives extracted from package comments.
type PackageDirectives struct {
	Package string
	Options map[string]string
}

func parsePackageDirectives(pkgs []*packages.Package) (PackageDirectives, error) {
	directives := PackageDirectives{
		Options: map[string]string{},
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if file.Doc == nil {
				continue
			}
			for _, comment := range file.Doc.List {
				if !strings.HasPrefix(comment.Text, "//protobuf:") {
					continue
				}
				input := strings.TrimPrefix(comment.Text, "//protobuf:")
				parts := strings.SplitN(input, " ", 2)
				//protobuf:package xyz.block.ftl.schema.v1
				//protobuf:option go_package="github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1;schemapb"
				switch parts[0] {
				case "package":
					directives.Package = parts[1]
				case "option":
					option := strings.SplitN(parts[1], "=", 2)
					if len(option) != 2 {
						return directives, genErrorf(comment.Pos(), "invalid option directive")
					}
					directives.Options[option[0]] = option[1]
				default:
					return directives, genErrorf(comment.Pos(), "unknown directive %q", parts[0])
				}
			}
		}
	}
	if directives.Package == "" {
		return directives, fmt.Errorf("missing //protobuf:package directive in package comments. Add a comment like '//protobuf:package xyz.block.ftl.schema.v1' to specify the protobuf package name")
	}
	return directives, nil
}
