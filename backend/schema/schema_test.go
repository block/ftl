package schema

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/errors"
)

var schema = MustValidate(&Schema{
	Modules: []*Module{
		{
			Name:     "todo",
			Comments: []string{"A comment"},
			Decls: []Decl{
				&Data{
					Name: "CreateRequest",
					Fields: []*Field{
						{Name: "name", Type: &Optional{Type: &Map{Key: &String{}, Value: &String{}}}},
					},
				},
				&Data{
					Name: "CreateResponse",
					Fields: []*Field{
						{Name: "name", Type: &Array{Element: &String{}}},
					},
				},
				&Data{
					Name: "DestroyRequest",
					Fields: []*Field{
						{Name: "name", Comments: []string{"A comment"}, Type: &String{}},
					},
				},
				&Data{
					Name: "DestroyResponse",
					Fields: []*Field{
						{Name: "name", Type: &String{}},
						{Name: "when", Type: &Time{}},
					},
				},
				&Verb{Name: "create",
					Request:  &DataRef{Module: "todo", Name: "CreateRequest"},
					Response: &DataRef{Module: "todo", Name: "CreateResponse"},
					Metadata: []Metadata{&MetadataCalls{Calls: []*VerbRef{{Module: "todo", Name: "destroy"}}}}},
				&Verb{Name: "destroy",
					Request:  &DataRef{Module: "todo", Name: "DestroyRequest"},
					Response: &DataRef{Module: "todo", Name: "DestroyResponse"},
					Metadata: []Metadata{
						&MetadataIngress{
							Type:   "ftl",
							Method: "GET",
							Path: []IngressPathComponent{
								&IngressPathLiteral{Text: "todo"},
								&IngressPathLiteral{Text: "destroy"},
								&IngressPathParameter{Name: "id"},
							},
						},
					},
				},
			},
		},
	},
})

func TestIndent(t *testing.T) {
	assert.Equal(t, "  a\n  b\n  c", indent("a\nb\nc"))
}

func TestSchemaString(t *testing.T) {
	expected := BuiltinsSource + `
// A comment
module todo {
  data CreateRequest {
    name {String: String}?
  }

  data CreateResponse {
    name [String]
  }

  data DestroyRequest {
    // A comment
    name String
  }

  data DestroyResponse {
    name String
    when Time
  }

  verb create(todo.CreateRequest) todo.CreateResponse  
      calls todo.destroy
      

  verb destroy(todo.DestroyRequest) todo.DestroyResponse  
      ingress ftl GET /todo/destroy/{id}
}
`
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(schema.String()))
}

func TestImports(t *testing.T) {
	input := `
	module test {
		data Data {
			ref other.Data
			ref another.Data
		}
	}
	`
	schema, err := ParseModuleString("", input)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, schema.Imports())
}

func TestVisit(t *testing.T) {
	expected := `
Module
  Data
    Field
      Optional
        Map
          String
          String
  Data
    Field
      Array
        String
  Data
    Field
      String
  Data
    Field
      String
    Field
      Time
  Verb
    DataRef
    DataRef
    MetadataCalls
      VerbRef
  Verb
    DataRef
    DataRef
    MetadataIngress
      IngressPathLiteral
      IngressPathLiteral
      IngressPathParameter
`
	actual := &strings.Builder{}
	i := 0
	// Modules[0] is always the builtins, which we skip.
	err := Visit(schema.Modules[1], func(n Node, next func() error) error {
		prefix := strings.Repeat(" ", i)
		tn := strings.TrimPrefix(fmt.Sprintf("%T", n), "*schema.")
		if strings.Contains(tn, "[") {
			tn = strings.TrimSuffix(strings.Split(tn, ".")[2], "]")
		}
		fmt.Fprintf(actual, "%s%s\n", prefix, tn)
		i += 2
		defer func() { i -= 2 }()
		return next()
	})
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(actual.String()), "%s", actual.String())
}

func TestParserRoundTrip(t *testing.T) {
	actual, err := ParseString("", schema.String())
	assert.NoError(t, err, "%s", schema.String())
	actual, err = Validate(actual)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(schema), Normalise(actual), "%s", schema.String())
}

func TestParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errors   []string
		expected *Schema
	}{
		{name: "Example",
			input: `
				module todo {
					data CreateListRequest {}
					data CreateListResponse {}

					// Create a new list
					verb createList(todo.CreateListRequest) CreateListResponse
						calls createList
				}
			`,
			expected: &Schema{
				Modules: []*Module{
					{
						Name: "todo",
						Decls: []Decl{
							&Data{Name: "CreateListRequest"},
							&Data{Name: "CreateListResponse"},
							&Verb{Name: "createList",
								Comments: []string{"Create a new list"},
								Request:  &DataRef{Module: "todo", Name: "CreateListRequest"},
								Response: &DataRef{Module: "todo", Name: "CreateListResponse"},
								Metadata: []Metadata{
									&MetadataCalls{Calls: []*VerbRef{{Module: "todo", Name: "createList"}}},
								},
							},
						},
					},
				},
			}},
		{name: "InvalidRequestRef",
			input: `module test { verb test(InvalidRequest) InvalidResponse}`,
			errors: []string{
				"1:25: reference to unknown data structure \"InvalidRequest\"",
				"1:41: reference to unknown data structure \"InvalidResponse\""}},
		{name: "InvalidDataRef",
			input: `module test { data Data { user user.User }}`,
			errors: []string{
				"1:32: reference to unknown data structure \"user.User\""}},
		{name: "InvalidMetadataSyntax",
			input: `module test { data Data {} calls }`,
			errors: []string{
				"1:28: unexpected token \"calls\" (expected \"}\")",
			},
		},
		{name: "InvalidDataMetadata",
			input: `module test { data Data {} calls verb }`,
			errors: []string{
				"1:28: metadata \"calls verb\" is not valid on data structures",
				"1:34: reference to unknown verb \"verb\"",
			}},
		{name: "KeywordAsName",
			input:  `module int { data String { name String } verb verb(String) String }`,
			errors: []string{"1:14: data structure name \"String\" is a reserved word"}},
		{name: "BuiltinRef",
			input: `module test { verb myIngress(HttpRequest) HttpResponse }`,
			expected: &Schema{
				Modules: []*Module{{
					Name: "test",
					Decls: []Decl{
						&Verb{
							Name:     "myIngress",
							Request:  &DataRef{Module: "builtin", Name: "HttpRequest"},
							Response: &DataRef{Module: "builtin", Name: "HttpResponse"},
						},
					},
				}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseString("", test.input)
			if test.errors != nil {
				assert.Error(t, err, "expected errors")
				actual := []string{}
				errs := errors.UnwrapAll(err)
				for _, err := range errs {
					if errors.Innermost(err) {
						actual = append(actual, err.Error())
					}
				}
				assert.Equal(t, test.errors, actual, test.input)
			} else {
				assert.NoError(t, err)
				actual = Normalise(actual)
				test.expected.Modules = append([]*Module{Builtins()}, test.expected.Modules...)
				assert.Equal(t, Normalise(test.expected), Normalise(actual), test.input)
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	input := `
// A comment
module todo {
  data CreateRequest {
    name {String: String}?
  }
  data CreateResponse {
    name [String]
  }
  data DestroyRequest {
    // A comment
    name String
  }
  data DestroyResponse {
    name String
	when Time
  }
  verb create(todo.CreateRequest) todo.CreateResponse
  	calls todo.destroy
  verb destroy(todo.DestroyRequest) todo.DestroyResponse
  	ingress ftl GET /todo/destroy/{id}
}
`
	actual, err := ParseModuleString("", input)
	assert.NoError(t, err)
	actual = Normalise(actual)
	assert.Equal(t, Normalise(schema.Modules[1]), Normalise(actual))
}
