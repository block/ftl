package console

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/schema/builder"
)

func TestVerbSchemaString(t *testing.T) {
	verb := &schema.Verb{
		Name:     "Echo",
		Request:  &schema.Ref{Module: "foo", Name: "EchoRequest"},
		Response: &schema.Ref{Module: "foo", Name: "EchoResponse"},
	}
	ingressVerb := &schema.Verb{
		Name:     "Ingress",
		Request:  &schema.Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []schema.Type{&schema.Unit{}, &schema.Unit{}, &schema.Unit{}}},
		Response: &schema.Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []schema.Type{&schema.String{}, &schema.String{}}},
		Metadata: []schema.Metadata{
			&schema.MetadataIngress{Type: "http", Method: "GET", Path: []schema.IngressPathComponent{&schema.IngressPathLiteral{Text: "test"}}},
		},
	}
	sch := builder.Schema(
		// TODO: Need a realm
		builder.Realm("").
			Module(
				schema.Builtins(),
				builder.Module("foo").
					Decl(
						verb,
						ingressVerb,
						&schema.Data{
							Name:       "EchoRequest",
							Visibility: schema.VisibilityScopeModule,
							Fields: []*schema.Field{
								{Name: "Name", Type: &schema.String{}},
								{Name: "Nested", Type: &schema.Ref{Module: "foo", Name: "Nested"}},
								{Name: "External", Type: &schema.Ref{Module: "bar", Name: "BarData"}},
								{Name: "Enum", Type: &schema.Ref{Module: "foo", Name: "Color"}},
							},
						},
						&schema.Data{
							Name:       "EchoResponse",
							Visibility: schema.VisibilityScopeModule,
							Fields: []*schema.Field{
								{Name: "Message", Type: &schema.String{}},
							},
						},
						&schema.Data{
							Name:       "Nested",
							Visibility: schema.VisibilityScopeModule,
							Fields: []*schema.Field{
								{Name: "Field", Type: &schema.String{}},
							},
						},
						&schema.Enum{
							Name:       "Color",
							Visibility: schema.VisibilityScopeModule,
							Type:       &schema.String{},
							Variants: []*schema.EnumVariant{
								{Name: "Red", Value: &schema.StringValue{Value: "Red"}},
								{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
								{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
							},
						},
					).
					MustBuild(),
				builder.Module("bar").
					Decl(
						&schema.Data{
							Name:       "BarData",
							Visibility: schema.VisibilityScopeModule,
							Fields: []*schema.Field{
								{Name: "Name", Type: &schema.String{}},
							},
						},
					).
					MustBuild(),
			).
			MustBuild()).
		MustBuild()

	expected := `export data EchoRequest {
  Name String
  Nested foo.Nested
  External bar.BarData
  Enum foo.Color
}

export data Nested {
  Field String
}

export data BarData {
  Name String
}

export enum Color: String {
  Red = "Red"
  Blue = "Blue"
  Green = "Green"
}

export data EchoResponse {
  Message String
}

verb Echo(foo.EchoRequest) foo.EchoResponse`

	schemaString, err := verbSchemaString(sch, verb)
	assert.NoError(t, err)
	assert.Equal(t, expected, schemaString)
}

func TestVerbSchemaStringIngress(t *testing.T) {
	verb := &schema.Verb{
		Name:     "Ingress",
		Request:  &schema.Ref{Module: "builtin", Name: "HttpRequest", TypeParameters: []schema.Type{&schema.Ref{Module: "foo", Name: "FooRequest"}, &schema.Unit{}, &schema.Unit{}}},
		Response: &schema.Ref{Module: "builtin", Name: "HttpResponse", TypeParameters: []schema.Type{&schema.Ref{Module: "foo", Name: "FooResponse"}, &schema.String{}}},
		Metadata: []schema.Metadata{
			&schema.MetadataIngress{Type: "http", Method: "POST", Path: []schema.IngressPathComponent{&schema.IngressPathLiteral{Text: "foo"}}},
		},
	}
	sch := builder.Schema(
		builder.Realm("").
			Module(
				builder.Module("foo").
					Decl(
						verb,
						&schema.Data{
							Name: "FooRequest",
							Fields: []*schema.Field{
								{Name: "Name", Type: &schema.String{}},
							},
						},
						&schema.Data{
							Name: "FooResponse",
							Fields: []*schema.Field{
								{Name: "Message", Type: &schema.String{}},
							},
						},
					).
					MustBuild()).
			MustBuild()).
		MustBuild()

	expected := `// HTTP request structure used for HTTP ingress verbs.
export data HttpRequest<Body, Path, Query> {
  method String
  path String
  pathParameters Path
  query Query
  headers {String: [String]}
  body Body
}

data FooRequest {
  Name String
}

// HTTP response structure used for HTTP ingress verbs.
export data HttpResponse<Body, Error> {
  status Int
  headers {String: [String]}
  // Either "body" or "error" must be present, not both.
  body Body?
  error Error?
}

data FooResponse {
  Message String
}

verb Ingress(builtin.HttpRequest<foo.FooRequest, Unit, Unit>) builtin.HttpResponse<foo.FooResponse, String>` + "  " + `
  +ingress http POST /foo`

	schemaString, err := verbSchemaString(sch, verb)
	assert.NoError(t, err)
	assert.Equal(t, expected, schemaString)
}
