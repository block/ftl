package main

import (
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/strcase"
)

var tmpl = template.Must(template.New("proto").
	Funcs(template.FuncMap{
		"typeof": func(t any) Kind { return Kind(reflect.Indirect(reflect.ValueOf(t)).Type().Name()) },
		"formatComment": func(comment string) string {
			if comment == "" {
				return ""
			}
			return "// " + strings.ReplaceAll(strings.TrimSpace(comment), "\n", "\n// ")
		},
		"toLowerCamel": strcase.ToLowerCamel,
		"toUpperCamel": strcase.ToUpperCamel,
		"toLowerSnake": strcase.ToLowerSnake,
		"toUpperSnake": strcase.ToUpperSnake,
		"trimPrefix":   strings.TrimPrefix,
	}).
	Parse(`// Code generated by go2proto. DO NOT EDIT.
syntax = "proto3";

package {{ .Package }};
{{ range .Imports }}
import "{{.}}";
{{- end}}
{{ range $name, $value := .Options }}
option {{ $name }} = {{ $value }};
{{- end }}
{{ range $decl := .OrderedDecls }}
{{- if eq (typeof $decl) "Message" }}
{{- .Comment | formatComment }}
message {{ .Name }} {
{{- range $name, $field := .Fields }}
  {{ if .Repeated }}repeated {{else if .Optional}}optional {{ end }}{{ .ProtoType }} {{ .Name | toLowerSnake }} = {{ .ID }};
{{- end }}
}
{{- else if eq (typeof $decl) "Enum" }}
{{- .Comment | formatComment }}
enum {{ .Name }} {
{{- range $value, $name := .ByValue }}
  {{ $name | toUpperSnake }} = {{ $value }};
{{- end }}
}
{{- else if eq (typeof $decl) "SumType" }}
{{ $sumtype := . }}
{{- .Comment | formatComment }}
message {{ .Name }} {
  oneof value {
{{- range $name, $id := .Variants }}
    {{ $name }} {{ $name | $sumtype.FieldName }} = {{ $id }};
{{- end }}
  }
}
{{- end }}
{{ end }}
`))

type RenderContext struct {
	PackageDirectives
	File
}

func render(out *os.File, directives PackageDirectives, file File) error {
	err := tmpl.Execute(out, RenderContext{
		PackageDirectives: directives,
		File:              file,
	})
	if err != nil {
		return errors.Wrap(err, "template error")
	}
	return nil
}
