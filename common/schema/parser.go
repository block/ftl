package schema

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/block/ftl/common/builderrors"
)

var (
	declUnion            = []Decl{&Data{}, &Verb{}, &Database{}, &Enum{}, &TypeAlias{}, &Config{}, &Secret{}, &Topic{}}
	nonOptionalTypeUnion = []Type{
		&Int{}, &Float{}, &String{}, &Bytes{}, &Bool{}, &Time{}, &Array{},
		&Map{}, &Any{}, &Unit{}, &Data{}, &TypeAlias{}, &Enum{},
		// Note: any types resolved by identifier (eg. "Any", "Unit", etc.) must
		// be prior to Ref.
		&Ref{},
	}
	metadataUnion = []Metadata{&MetadataCalls{}, &MetadataConfig{}, &MetadataIngress{}, &MetadataCronJob{},
		&MetadataDatabases{}, &MetadataAlias{}, &MetadataRetry{}, &MetadataSecrets{}, &MetadataSubscriber{},
		&MetadataTypeMap{}, &MetadataEncoding{}, &MetadataPublisher{}, &MetadataSQLMigration{}, &MetadataArtefact{},
		&MetadataPartitions{}, &MetadataSQLQuery{}, &MetadataSQLColumn{}, &MetadataGenerated{}, &MetadataGit{}, &MetadataFixture{},
		&MetadataTransaction{}, &MetadataEgress{}, &MetadataImage{}}
	ingressUnion = []IngressPathComponent{&IngressPathLiteral{}, &IngressPathParameter{}}
	valueUnion   = []Value{&StringValue{}, &IntValue{}, &TypeValue{}}

	Lexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "EOL", Pattern: `[\r\n]`},
		{Name: "SHA256", Pattern: `\b[A-za-z0-9]{64}\b`},
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
		{Name: "Comment", Pattern: `//.*`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
		{Name: "Punct", Pattern: `[%/\-\_:[\]{}<>()*+?.,\\^$|#~!\'@=]`},
	})

	commonParserOptions = []participle.Option{
		participle.Lexer(Lexer),
		participle.Elide("Whitespace", "EOL"),
		participle.Unquote(),
		// Increase lookahead to allow comments to be attached to the next token.
		participle.UseLookahead(participle.MaxLookahead),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
		participle.Union(metadataUnion...),
		participle.Union(ingressUnion...),
		participle.Union(declUnion...),
		participle.Union(valueUnion...),
	}

	// Parser options for every parser _except_ the type parser.
	parserOptions = append(commonParserOptions, participle.ParseTypeWith(ParseTypeWithLexer))

	parser       = participle.MustBuild[Schema](parserOptions...)
	moduleParser = participle.MustBuild[Module](parserOptions...)
	typeParser   = participle.MustBuild[typeParserGrammar](append(commonParserOptions, participle.Union(nonOptionalTypeUnion...))...)
	refParser    = participle.MustBuild[Ref](parserOptions...)
)

type Position struct {
	Filename string `protobuf:"1"`
	Offset   int    `parser:"" protobuf:"-"`
	Line     int    `protobuf:"2"`
	Column   int    `protobuf:"3"`
}

func (p Position) String() string {
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}

func (p Position) ToErrorPos() builderrors.Position {
	return p.ToErrorPosWithEnd(p.Column)
}

func (p Position) ToErrorPosWithEnd(endColumn int) builderrors.Position {
	return builderrors.Position{
		Filename:    p.Filename,
		Line:        p.Line,
		StartColumn: p.Column,
		EndColumn:   endColumn,
	}
}

// A Node in the schema grammar.
//
//sumtype:decl
type Node interface {
	String() string
	Position() Position
	// schemaChildren returns the children of this node.
	schemaChildren() []Node
}

// Kind of a Symbol in the schema.
type Kind int

const (
	KindData Kind = iota
	KindInt
	KindFloat
	KindString
	KindBool
	KindBytes
	KindOptional
	KindMap
	KindArray
	KindRef
	KindTime
	KindAny
	KindUnit
	KindEnum
	KindTypeAlias
)

func (k Kind) String() string {
	switch k {
	case KindData:
		return "data"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindBytes:
		return "bytes"
	case KindOptional:
		return "optional"
	case KindMap:
		return "map"
	case KindArray:
		return "array"
	case KindRef:
		return "ref"
	case KindTime:
		return "time"
	case KindAny:
		return "any"
	case KindUnit:
		return "unit"
	case KindEnum:
		return "enum"
	case KindTypeAlias:
		return "type alias"
	default:
		return "unknown"
	}
}

// Type represents a Type Node in the schema grammar.
//
//sumtype:decl
type Type interface {
	Node
	// Equal returns true if this type is equal to another type.
	Equal(other Type) bool
	// Kind of the Symbol.
	Kind() Kind
	// schemaType is a marker to ensure that all types implement the Type interface.
	schemaType()
}

// Metadata represents a metadata Node in the schema grammar.
//
//sumtype:decl
type Metadata interface {
	Node
	schemaMetadata()
}

// Value represents a value Node in the schema grammar.
//
//sumtype:decl
type Value interface {
	Node
	Equal(other Value) bool
	GetValue() any
	schemaValueType() Type
}

// Symbol represents a symbol in the schema grammar.
//
// A Symbol is a named type that can be referenced by other types. This includes
// user defined data types such as data structures and enums, and builtin types.
//
//sumtype:decl
type Symbol interface {
	Node
	schemaSymbol()
}

// A Named symbol in the grammar.
type Named interface {
	Symbol
	GetName() string
}

// Decl represents user-defined data types in the schema grammar.
//
//sumtype:decl
type Decl interface {
	Symbol
	GetName() string
	// IsGenerated returns true if the Decl is in the schema but not in the source code.
	IsGenerated() bool
	GetVisibility() Visibility
	schemaDecl()
}

// We have a separate parser for types because Participle doesn't support left
// recursion and "Type = Type ? | Int | String ..." is left recursive.
type typeParserGrammar struct {
	Type     Type `parser:"@@"`
	Optional bool `parser:"@'?'?"`
}

// ParseType parses the string representation of a type.
func ParseType(filename, input string) (Type, error) {
	typ, err := typeParser.ParseString(filename, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if typ.Optional {
		return &Optional{Type: typ.Type}, nil
	}
	return typ.Type, nil
}

func ParseTypeWithLexer(pl *lexer.PeekingLexer) (Type, error) {
	typ, err := typeParser.ParseFromLexer(pl, participle.AllowTrailing(true))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if typ.Optional {
		return &Optional{Type: typ.Type}, nil
	}
	return typ.Type, nil
}

func ParseString(filename, input string) (*Schema, error) {
	mod, err := parser.ParseString(filename, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(mod.Validate())
}

func ParseModuleString(filename, input string) (*Module, error) {
	mod, err := moduleParser.ParseString(filename, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, errors.WithStack(mod.Validate())
}

func Parse(filename string, r io.Reader) (*Schema, error) {
	mod, err := parser.Parse(filename, r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(mod.Validate())
}

func ParseModule(filename string, r io.Reader) (*Module, error) {
	mod, err := moduleParser.Parse(filename, r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, errors.WithStack(mod.Validate())
}

// EBNF grammar for the FTL schema.
func EBNF() string {
	return parser.String()
}
