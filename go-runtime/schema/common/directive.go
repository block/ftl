package common

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl/common/cron"
	"github.com/block/ftl/common/schema"
)

// This file contains a parser for Go FTL directives.

type directiveWrapper struct {
	Directive Directive `parser:"'ftl' ':' @@"`
}

// Directive is a directive in a Go FTL module, e.g. //ftl:ingress http GET /foo/bar
//
//sumtype:decl
type Directive interface {
	SetPosition(pos token.Pos)
	GetPosition() token.Pos
	GetTypeName() string
	// MustAnnotate returns the AST nodes that can be annotated by this directive.
	MustAnnotate() []ast.Node

	directive()
}

type Exportable interface {
	IsExported() bool
}

type DirectiveVerb struct {
	Pos token.Pos

	Verb   bool `parser:"@'verb'"`
	Export bool `parser:"@'export'?"`
}

func (*DirectiveVerb) directive() {}
func (d *DirectiveVerb) String() string {
	if d.Export {
		return "ftl:verb export"
	}
	return "ftl:verb"
}
func (d *DirectiveVerb) IsExported() bool {
	return d.Export
}
func (*DirectiveVerb) GetTypeName() string { return "verb" }
func (d *DirectiveVerb) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveVerb) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveVerb) MustAnnotate() []ast.Node { return []ast.Node{&ast.FuncDecl{}} }

type DirectiveData struct {
	Pos token.Pos

	Data   bool `parser:"@'data'"`
	Export bool `parser:"@'export'?"`
}

func (*DirectiveData) directive() {}
func (d *DirectiveData) String() string {
	if d.Export {
		return "ftl:data export"
	}
	return "ftl:data"
}
func (d *DirectiveData) IsExported() bool {
	return d.Export
}
func (*DirectiveData) GetTypeName() string { return "data" }
func (d *DirectiveData) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveData) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveData) MustAnnotate() []ast.Node { return []ast.Node{&ast.GenDecl{}} }

type DirectiveEnum struct {
	Pos token.Pos

	Enum   bool `parser:"@'enum'"`
	Export bool `parser:"@'export'?"`
}

func (*DirectiveEnum) directive() {}
func (d *DirectiveEnum) String() string {
	if d.Export {
		return "ftl:enum export"
	}
	return "ftl:enum"
}
func (d *DirectiveEnum) IsExported() bool {
	return d.Export
}
func (*DirectiveEnum) GetTypeName() string { return "enum" }
func (d *DirectiveEnum) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveEnum) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveEnum) MustAnnotate() []ast.Node { return []ast.Node{&ast.GenDecl{}} }

type DirectiveTypeAlias struct {
	Pos token.Pos

	TypeAlias bool `parser:"@'typealias'"`
	Export    bool `parser:"@'export'?"`
}

func (*DirectiveTypeAlias) directive() {}
func (d *DirectiveTypeAlias) String() string {
	if d.Export {
		return "ftl:typealias export"
	}
	return "ftl:typealias"
}
func (d *DirectiveTypeAlias) IsExported() bool {
	return d.Export
}
func (*DirectiveTypeAlias) GetTypeName() string { return "typealias" }
func (d *DirectiveTypeAlias) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveTypeAlias) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveTypeAlias) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.GenDecl{}}
}

type DirectiveIngress struct {
	Pos token.Pos

	Type   string                        `parser:"'ingress' @('http')?"`
	Method string                        `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')"`
	Path   []schema.IngressPathComponent `parser:"('/' @@)+"`
}

func (*DirectiveIngress) directive() {}
func (d *DirectiveIngress) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "ftl:ingress %s", d.Method)
	for _, p := range d.Path {
		fmt.Fprintf(w, "/%s", p)
	}
	return w.String()
}
func (d *DirectiveIngress) IsExported() bool {
	return true
}
func (*DirectiveIngress) GetTypeName() string { return "ingress" }
func (d *DirectiveIngress) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveIngress) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveIngress) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

type DirectiveCronJob struct {
	Pos token.Pos

	Cron cron.Pattern `parser:"'cron' @@"`
}

func (*DirectiveCronJob) directive() {}

func (d *DirectiveCronJob) String() string {
	return fmt.Sprintf("cron %s", d.Cron)
}
func (d *DirectiveCronJob) IsExported() bool {
	return false
}
func (*DirectiveCronJob) GetTypeName() string { return "cron" }
func (d *DirectiveCronJob) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveCronJob) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveCronJob) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

type DirectiveFixture struct {
	Pos token.Pos

	Manual bool `parser:"'fixture' @'manual'?"`
}

func (*DirectiveFixture) directive() {}

func (d *DirectiveFixture) String() string {
	if d.Manual {
		return "fixture manual"
	}
	return "fixture"
}
func (d *DirectiveFixture) IsExported() bool {
	return false
}
func (*DirectiveFixture) GetTypeName() string { return "fixture" }
func (d *DirectiveFixture) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveFixture) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveFixture) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

type DirectiveRetry struct {
	Pos token.Pos

	Count      *int        `parser:"'retry' (@Number Whitespace)?"`
	MinBackoff string      `parser:"@(Number (?! Whitespace) Ident)?"`
	MaxBackoff string      `parser:"@(Number (?! Whitespace) Ident)?"`
	Catch      *schema.Ref `parser:"('catch' @@)?"`
}

func (*DirectiveRetry) directive() {}

func (d *DirectiveRetry) String() string {
	components := []string{"retry"}
	if d.Count != nil {
		components = append(components, strconv.Itoa(*d.Count))
	}
	components = append(components, d.MinBackoff)
	if len(d.MaxBackoff) > 0 {
		components = append(components, d.MaxBackoff)
	}
	if d.Catch != nil {
		components = append(components, fmt.Sprintf("catch %v", d.Catch))
	}
	return strings.Join(components, " ")
}
func (*DirectiveRetry) GetTypeName() string { return "retry" }
func (d *DirectiveRetry) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveRetry) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveRetry) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}, &ast.GenDecl{}}
}

// DirectiveTopic is used to configure options for a topic.
type DirectiveTopic struct {
	Pos token.Pos

	Export     bool `parser:"'topic' @'export'?"`
	Partitions int  `parser:"'partitions' '=' @Number"`
}

func (*DirectiveTopic) directive() {}

func (d *DirectiveTopic) String() string {
	components := []string{"topic"}
	if d.Export {
		components = append(components, "export")
	}
	components = append(components, "partitions="+strconv.Itoa(d.Partitions))
	return strings.Join(components, " ")
}

func (*DirectiveTopic) GetTypeName() string { return "topic" }
func (d *DirectiveTopic) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveTopic) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveTopic) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}, &ast.GenDecl{}}
}

func (d *DirectiveTopic) IsExported() bool {
	return d.Export
}

// DirectiveSubscriber is used to subscribe a sink to a subscription
type DirectiveSubscriber struct {
	Pos token.Pos

	Topic      *schema.Ref        `parser:"'subscribe' @@"`
	FromOffset *schema.FromOffset `parser:"'from' '='@('beginning'|'latest')"`
	DeadLetter bool               `parser:"@'deadletter'?"`
}

func (*DirectiveSubscriber) directive() {}

func (d *DirectiveSubscriber) String() string {
	components := []string{"subscribe", d.Topic.String()}
	components = append(components, "from="+d.FromOffset.String())
	if d.DeadLetter {
		components = append(components, "deadletter")
	}
	return strings.Join(components, " ")
}
func (*DirectiveSubscriber) GetTypeName() string { return "subscribe" }
func (d *DirectiveSubscriber) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveSubscriber) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveSubscriber) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

// DirectiveExport is used on declarations that don't include export in other directives.
type DirectiveExport struct {
	Pos token.Pos

	Export bool `parser:"@'export'"`
}

func (*DirectiveExport) directive() {}

func (d *DirectiveExport) String() string {
	return "export"
}
func (*DirectiveExport) GetTypeName() string { return "export" }
func (d *DirectiveExport) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveExport) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveExport) MustAnnotate() []ast.Node { return []ast.Node{&ast.GenDecl{}} }
func (d *DirectiveExport) IsExported() bool {
	return d.Export
}

// DirectiveTypeMap is used to declare a native type to deserialize to in a given runtime.
type DirectiveTypeMap struct {
	Pos token.Pos

	Runtime    string `parser:"'typemap' @('go' | 'kotlin' | 'java')"`
	NativeName string `parser:"@String"`
}

func (*DirectiveTypeMap) directive() {}

func (d *DirectiveTypeMap) String() string {
	return fmt.Sprintf("typemap %s %q", d.Runtime, d.NativeName)
}
func (*DirectiveTypeMap) GetTypeName() string { return "typemap" }
func (d *DirectiveTypeMap) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveTypeMap) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveTypeMap) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.GenDecl{}}
}

// DirectiveEncoding can be used to enable custom encoding behavior.
type DirectiveEncoding struct {
	Pos token.Pos

	Type    string `parser:"'encoding' @('json')?"`
	Lenient bool   `parser:"@'lenient'"`
}

func (*DirectiveEncoding) directive() {}

func (d *DirectiveEncoding) String() string {
	components := []string{"encoding"}
	if d.Type != "" {
		components = append(components, d.Type)
	}
	if d.Lenient {
		components = append(components, "lenient")
	}
	return strings.Join(components, " ")
}
func (*DirectiveEncoding) GetTypeName() string { return "encoding" }
func (d *DirectiveEncoding) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveEncoding) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveEncoding) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

type DirectiveDatabase struct {
	Pos token.Pos

	Engine string `parser:"'database' @('postgres' | 'mysql')"`
	Name   string `parser:"@Ident"`
}

func (*DirectiveDatabase) directive() {}

func (*DirectiveDatabase) GetTypeName() string { return "database" }
func (d *DirectiveDatabase) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveDatabase) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveDatabase) MustAnnotate() []ast.Node { return []ast.Node{&ast.GenDecl{}} }

type DirectiveTransaction struct {
	Pos token.Pos

	Transaction bool `parser:"'transaction'"`
}

func (*DirectiveTransaction) directive() {}

func (d *DirectiveTransaction) String() string {
	return "transaction"
}
func (d *DirectiveTransaction) IsExported() bool {
	return false
}
func (*DirectiveTransaction) GetTypeName() string { return "transaction" }
func (d *DirectiveTransaction) SetPosition(pos token.Pos) {
	d.Pos = pos
}
func (d *DirectiveTransaction) GetPosition() token.Pos {
	return d.Pos
}
func (*DirectiveTransaction) MustAnnotate() []ast.Node {
	return []ast.Node{&ast.FuncDecl{}}
}

var DirectiveParser = participle.MustBuild[directiveWrapper](
	participle.Lexer(schema.Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[Directive](&DirectiveVerb{}, &DirectiveData{}, &DirectiveEnum{}, &DirectiveTypeAlias{},
		&DirectiveIngress{}, &DirectiveCronJob{}, &DirectiveRetry{}, &DirectiveSubscriber{}, &DirectiveExport{},
		&DirectiveTypeMap{}, &DirectiveEncoding{}, &DirectiveTopic{}, &DirectiveDatabase{}, &DirectiveFixture{},
		&DirectiveTransaction{}),
	participle.Union[schema.IngressPathComponent](&schema.IngressPathLiteral{}, &schema.IngressPathParameter{}),
	participle.ParseTypeWith(schema.ParseTypeWithLexer),
)

func ParseDirectives(pass *analysis.Pass, node ast.Node, docs *ast.CommentGroup) []Directive {
	if docs == nil {
		return nil
	}
	var directives []Directive
	for _, line := range docs.List {
		if !strings.HasPrefix(line.Text, "//ftl:") {
			continue
		}
		pos := pass.Fset.Position(line.Pos())
		// TODO: We need to adjust position information embedded in the schema.
		directive, err := DirectiveParser.ParseString(pos.Filename, line.Text[2:])
		file := pass.Fset.File(node.Pos())
		startPos := file.Pos(file.Offset(line.Pos()) + 2)
		if err != nil {
			// Adjust the Participle-reported position relative to the AST node.
			var perr participle.Error
			if errors.As(err, &perr) {
				errorfAtPos(pass, startPos, file.Pos(file.Offset(line.End())), "%s", perr.Message())
			} else {
				Wrapf(pass, node, err, "")
			}
			return nil
		}
		directive.Directive.SetPosition(startPos)
		directives = append(directives, directive.Directive)
	}
	return directives
}
