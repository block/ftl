package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/block/ftl/common/strcase"
)

var _ fmt.Stringer

var go2protoTmpl = template.Must(template.New("go2proto.to.go.tmpl").
	Funcs(template.FuncMap{
		"typeof": func(t any) Kind { return Kind(reflect.Indirect(reflect.ValueOf(t)).Type().Name()) },
		// Return true if the type is a builtin proto type.
		"isBuiltin": func(t Field) bool {
			switch t.OriginType {
			case "int", "uint", "int32", "int64", "uint32", "uint64", "float32", "float64", "bool", "string":
				return true
			}
			return false
		},
		"protoName": protoName,
		"goProtoImport": func(g Go2ProtoContext) (string, error) {
			unquoted, err := strconv.Unquote(g.Options["go_package"])
			if err != nil {
				return "", fmt.Errorf("go_package must be a quoted string: %w", err)
			}
			parts := strings.Split(unquoted, ";")
			return parts[0], nil
		},
		"sumTypeVariantName": func(s string, v string) string {
			return protoName(strings.TrimPrefix(v, s))
		},
		"toLower":      strings.ToLower,
		"toUpper":      strings.ToUpper,
		"toLowerCamel": strcase.ToLowerCamel,
		"toUpperCamel": strcase.ToUpperCamel,
		"toLowerSnake": strcase.ToLowerSnake,
		"toUpperSnake": strcase.ToUpperSnake,
		"trimPrefix":   strings.TrimPrefix,
	}).
	Parse(`// Code generated by go2proto. DO NOT EDIT.

package {{ .GoPackage }}

import "fmt"
import destpb "{{ . | goProtoImport }}"
import "google.golang.org/protobuf/proto"
import "google.golang.org/protobuf/types/known/timestamppb"
import "google.golang.org/protobuf/types/known/durationpb"
{{range $gimport := .GoImports }}
import "{{ $gimport }}"
{{- end}}


var _ fmt.Stringer
var _ = timestamppb.Timestamp{}
var _ = durationpb.Duration{}

// protoSlice converts a slice of values to a slice of protobuf values.
func protoSlice[P any, T interface{ ToProto() P }](values []T) []P {
	out := make([]P, len(values))
	for i, v := range values {
		out[i] = v.ToProto()
	}
	return out
}

func protoMust[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func sliceMap[T any, U any](values []T, f func(T) U) []U {
	out := make([]U, len(values))
	for i, v := range values {
		out[i] = f(v)
	}
	return out
}

func sliceMapErr[T any, U any](values []T, f func(T) (U, error)) ([]U, error) {
	var err error
	out := make([]U, len(values))
	for i, v := range values {
		if out[i], err = f(v);  err != nil {
			return nil, err
		}
	}
	return out, nil
}

func orZero[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func ptr[T any, O any](v *O, o T) *T {
	if v == nil {
		return nil
	}
	return &o
}

func fromPtr[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

{{range $decl := .OrderedDecls }}
{{- if eq (typeof $decl) "Message" }}
func (x *{{ .Name }}) ToProto() *destpb.{{ .Name }} {
	if x == nil {
		return nil
	}
	return &destpb.{{ .Name }}{
{{- range $field := .Fields }}
{{- if . | isBuiltin }}
{{- if $field.Optional}}
		{{ $field.EscapedName }}: proto.{{ $field.ProtoGoType | toUpperCamel }}({{ $field.ProtoGoType }}({{if $field.Pointer}}*{{end}}x.{{ $field.Name }})),
{{- else if .Repeated}}
		{{ $field.EscapedName }}: sliceMap(x.{{ $field.Name }}, func(v {{ $field.OriginType }}) {{ $field.ProtoGoType }} { return {{ $field.ProtoGoType }}(v) }),
{{- else }}
		{{ $field.EscapedName }}: {{ $field.ProtoGoType }}(x.{{ $field.Name }}),
{{- end}}
{{- else if eq $field.ProtoType "google.protobuf.Timestamp" }}
		{{ $field.EscapedName }}: timestamppb.New(x.{{ $field.Name }}),
{{- else if eq $field.ProtoType "google.protobuf.Duration" }}
		{{ $field.EscapedName }}: durationpb.New(x.{{ $field.Name }}),
{{- else if eq .Kind "Message" }}
{{- if .Repeated }}
		{{ $field.EscapedName }}: protoSlice[*destpb.{{ .ProtoGoType }}](x.{{ $field.Name }}),
{{- else}}
		{{ $field.EscapedName }}: x.{{ $field.Name }}.ToProto(),
{{- end}}
{{- else if eq .Kind "Enum" }}
{{- if .Repeated }}
		{{ $field.EscapedName }}: protoSlice[destpb.{{ .Type }}](x.{{ $field.Name }}),
{{- else}}
		{{ $field.EscapedName }}: x.{{ $field.Name }}.ToProto(),
{{- end}}
{{- else if eq .Kind "SumType" }}
{{- if .Repeated }}
		{{ $field.EscapedName }}: sliceMap(x.{{ $field.Name }}, {{$field.OriginType}}ToProto),
{{- else}}
		{{ $field.EscapedName }}: {{ $field.OriginType }}ToProto(x.{{ $field.Name }}),
{{- end}}
{{- else if eq $field.Kind "BinaryMarshaler" }}
		{{ $field.EscapedName }}: protoMust(x.{{ $field.Name }}.MarshalBinary()),
{{- else if eq $field.Kind "TextMarshaler" }}
		{{ $field.EscapedName }}: string(protoMust(x.{{ $field.Name }}.MarshalText())),
{{- else }}
		{{ $field.EscapedName }}: ??, // x.{{ $field.Name }}.ToProto() // Unknown type {{ $field.OriginType }} of kind {{ $field.Kind }}
{{- end}}
{{- end}}
	}
}

func {{ .Name }}FromProto(v *destpb.{{ .Name }}) (out *{{ .Name }}, err error) {
	if v == nil {
		return nil, nil
	}

	out = &{{ .Name }}{}

{{- range $field := .Fields }}
{{- if . | isBuiltin }}
{{- if $field.Optional}}
{{- if $field.Pointer}}
	out.{{ $field.Name }} = ptr(v.{{ $field.EscapedName }}, {{ $field.OriginType }}(orZero(v.{{ $field.EscapedName }})))
{{- else}}
	out.{{ $field.Name }} = {{ $field.OriginType }}(orZero(v.{{ $field.EscapedName }}))
{{- end}}
{{- else if .Repeated}}
	out.{{ $field.Name }} = sliceMap(v.{{ $field.EscapedName }}, func(v {{ $field.ProtoGoType }}) {{ $field.OriginType }} { return {{ $field.OriginType }}(v) })
{{- else }}
	out.{{ $field.Name }} = {{ $field.OriginType }}(v.{{ $field.EscapedName }})
{{- end}}
{{- else if eq $field.ProtoType "google.protobuf.Timestamp" }}
	out.{{ $field.Name }} = v.{{ $field.EscapedName }}.AsTime()
{{- else if eq $field.ProtoType "google.protobuf.Duration" }}
	out.{{ $field.Name }} = v.{{ $field.EscapedName }}.AsDuration()
{{- else if eq .Kind "Message" }}
{{- if .Repeated }}
	if out.{{ $field.Name }}, err = sliceMapErr(v.{{ $field.EscapedName }}, {{$field.OriginType}}FromProto); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else}}
{{- if $field.Pointer}}
	if out.{{ $field.Name }}, err = {{ $field.OriginType }}FromProto(v.{{ $field.EscapedName }}); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else}}
	if field{{ $field.Name }}, err := {{ $field.OriginType }}FromProto(v.{{ $field.EscapedName }}); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	} else {
		out.{{ $field.Name }} = fromPtr(field{{ $field.Name }})
	}
{{- end}}
{{- end}}
{{- else if eq .Kind "Enum" }}
{{- if .Repeated }}
	if out.{{ $field.Name }}, err = sliceMapErr(v.{{ $field.EscapedName }}, {{$field.OriginType}}FromProto); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else}}
	if out.{{ $field.Name }}, err = {{ $field.OriginType }}FromProto(v.{{ $field.EscapedName }}); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- end}}
{{- else if eq .Kind "SumType" }}
{{- if .Repeated }}
	if out.{{ $field.Name }}, err = sliceMapErr(v.{{ $field.EscapedName }}, {{$field.OriginType}}FromProto); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else}}
	if out.{{ $field.Name }}, err = {{ $field.OriginType }}FromProto(v.{{ $field.EscapedName }}); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- end}}
{{- else if eq $field.Kind "BinaryMarshaler" }}
{{- if $field.Pointer}}
	out.{{ $field.Name }} = new({{ $field.OriginType }})
{{- end}}
	if err = out.{{ $field.Name }}.UnmarshalBinary(v.{{ $field.EscapedName }}); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else if eq $field.Kind "TextMarshaler" }}
{{- if $field.Pointer}}
	out.{{ $field.Name }} = new({{ $field.OriginType }})
{{- end}}
	if err = out.{{ $field.Name }}.UnmarshalText([]byte(v.{{ $field.EscapedName }})); err != nil {
		return nil, fmt.Errorf("{{ $field.Name }}: %w", err)
	}
{{- else }}
	out.{{ $field.Name }} = ??, // v.{{ $field.EscapedName }}.ToProto() // Unknown type {{ $field.OriginType }} of kind {{ $field.Kind }}
{{- end}}
{{- end}}
	return out, nil
}

{{- else if eq (typeof $decl) "Enum" }}
func (x {{ .Name }}) ToProto() destpb.{{ .Name }} {
	return destpb.{{ .Name }}(x)
}

func {{ .Name }}FromProto(v destpb.{{ .Name }}) ({{ .Name }}, error) {
	// TODO: Check if the value is valid.
	return {{ .Name }}(v), nil
}
{{- else if eq (typeof $decl) "SumType" }}
{{- $sumtype := . }}
// {{ .Name }}ToProto converts a {{ .Name }} sum type to a protobuf message.
func {{ .Name }}ToProto(value {{ .Name }}) *destpb.{{ .Name }} {
	switch value := value.(type) {
	case nil:
		return nil
	{{- range $variant, $id := .Variants }}
	case *{{ $variant }}:
		return &destpb.{{ $sumtype.Name }}{
			Value: &destpb.{{ $sumtype.Name | toUpperCamel }}_{{ sumTypeVariantName $sumtype.Name $variant }}{value.ToProto()},
		}
	{{- end }}
	default:
		panic(fmt.Sprintf("unknown variant: %T", value))
	}
}

func {{ $sumtype.Name }}FromProto(v *destpb.{{ $sumtype.Name }}) ({{ $sumtype.Name }}, error) {
	if v == nil {
		return nil, nil
	}
	switch v.Value.(type) {
	{{- range $variant, $id := .Variants }}
	case *destpb.{{ $sumtype.Name | toUpperCamel }}_{{ sumTypeVariantName $sumtype.Name $variant }}:
		return {{ $variant }}FromProto(v.Get{{ sumTypeVariantName $sumtype.Name $variant }}())
	{{- end }}
	default:
		panic(fmt.Sprintf("unknown variant: %T", v.Value))
	}
}
{{- end}}
{{ end}}
		`))

type Go2ProtoContext struct {
	PackageDirectives
	File
	GoImports []string
}

func renderToProto(out *os.File, directives PackageDirectives, file File, goImports []string) error {
	err := go2protoTmpl.Execute(out, Go2ProtoContext{
		PackageDirectives: directives,
		File:              file,
		GoImports:         goImports,
	})
	if err != nil {
		return fmt.Errorf("template error: %w", err)
	}
	return nil
}
