package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:ingress http GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface{ directive() }

type exportable interface {
	IsExported() bool
}

type directiveVerb struct {
	Pos lexer.Position

	Verb   bool `parser:"@'verb'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveVerb) directive() {}
func (d *directiveVerb) String() string {
	if d.Export {
		return "ftl:verb export"
	}
	return "ftl:verb"
}
func (d *directiveVerb) IsExported() bool {
	return d.Export
}

type directiveData struct {
	Pos lexer.Position

	Data   bool `parser:"@'data'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveData) directive() {}
func (d *directiveData) String() string {
	if d.Export {
		return "ftl:data export"
	}
	return "ftl:data"
}
func (d *directiveData) IsExported() bool {
	return d.Export
}

type directiveEnum struct {
	Pos lexer.Position

	Enum   bool `parser:"@'enum'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveEnum) directive() {}
func (d *directiveEnum) String() string {
	if d.Export {
		return "ftl:enum export"
	}
	return "ftl:enum"
}
func (d *directiveEnum) IsExported() bool {
	return d.Export
}

type directiveIngress struct {
	Pos schema.Position

	Type   string                        `parser:"'ingress' @('http')?"`
	Method string                        `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')"`
	Path   []schema.IngressPathComponent `parser:"('/' @@)+"`
}

func (*directiveIngress) directive() {}
func (d *directiveIngress) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "ftl:ingress %s", d.Method)
	for _, p := range d.Path {
		fmt.Fprintf(w, "/%s", p)
	}
	return w.String()
}

type directiveCronJob struct {
	Pos schema.Position

	Cron string `parser:"'cron' Whitespace @((' ' | Number | '-' | '/' | '*' | ',')+)"`
}

func (*directiveCronJob) directive() {}

func (d *directiveCronJob) String() string {
	return fmt.Sprintf("cron %s", d.Cron)
}

var directiveParser = participle.MustBuild[directiveWrapper](
	participle.Lexer(schema.Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[directive](&directiveVerb{}, &directiveData{}, &directiveEnum{}, &directiveIngress{}, &directiveCronJob{}),
	participle.Union[schema.IngressPathComponent](&schema.IngressPathLiteral{}, &schema.IngressPathParameter{}),
)

func parseDirectives(node ast.Node, fset *token.FileSet, docs *ast.CommentGroup) ([]directive, *schema.Error) {
	if docs == nil {
		return nil, nil
	}
	directives := []directive{}
	for _, line := range docs.List {
		if !strings.HasPrefix(line.Text, "//ftl:") {
			continue
		}
		pos := fset.Position(line.Pos())
		// TODO: We need to adjust position information embedded in the schema.
		directive, err := directiveParser.ParseString(pos.Filename, line.Text[2:])
		if err != nil {
			// Adjust the Participle-reported position relative to the AST node.
			var scerr *schema.Error
			var perr participle.Error
			if errors.As(err, &perr) {
				ppos := schema.Position{}
				ppos.Filename = pos.Filename
				ppos.Column += pos.Column + 2
				ppos.Line = pos.Line
				scerr = schema.Errorf(ppos, ppos.Column, "%s", perr.Message())
			} else {
				scerr = wrapf(node, err, "")
			}
			return nil, scerr
		}
		directives = append(directives, directive.Directive)
	}
	return directives, nil
}
