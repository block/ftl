package buildengine

import (
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
)

func TestGenerateGoModule(t *testing.T) {
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "other", Decls: []schema.Decl{
				&schema.Enum{
					Name: "Color",
					Type: &schema.String{},
					Variants: []*schema.EnumVariant{
						{Name: "Red", Value: &schema.StringValue{Value: "Red"}},
						{Name: "Blue", Value: &schema.StringValue{Value: "Blue"}},
						{Name: "Green", Value: &schema.StringValue{Value: "Green"}},
					},
				},
				&schema.Enum{
					Name: "ColorInt",
					Type: &schema.Int{},
					Variants: []*schema.EnumVariant{
						{Name: "RedInt", Value: &schema.IntValue{Value: 0}},
						{Name: "BlueInt", Value: &schema.IntValue{Value: 1}},
						{Name: "GreenInt", Value: &schema.IntValue{Value: 2}},
					},
				},
				&schema.Data{Name: "EchoRequest"},
				&schema.Data{Name: "EchoResponse"},
				&schema.Verb{
					Name:     "echo",
					Request:  &schema.Ref{Name: "EchoRequest"},
					Response: &schema.Ref{Name: "EchoResponse"},
				},
			}},
			{Name: "test"},
		},
	}
	expected := `// Code generated by FTL. DO NOT EDIT.

//ftl:module other
package other

import (
  "context"
)

var _ = context.Background

//ftl:enum
type Color string
const (
  Red Color = "Red"
  Blue Color = "Blue"
  Green Color = "Green"
)

//ftl:enum
type ColorInt int
const (
  RedInt ColorInt = 0
  BlueInt ColorInt = 1
  GreenInt ColorInt = 2
)

type EchoRequest struct {
}

type EchoResponse struct {
}

//ftl:verb
func Echo(context.Context, EchoRequest) (EchoResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/ftl.Call()")
}
`
	bctx := buildContext{
		moduleDir: "testdata/modules/another",
		buildDir:  "_ftl",
		sch:       sch,
	}
	testBuild(t, bctx, []assertion{
		assertGeneratedModule("go/modules/other/external_module.go", expected),
	})
}
