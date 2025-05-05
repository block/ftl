package builder

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
)

func TestBuildSchemaError(t *testing.T) {
	builder := Module("service").Decl(&schema.Config{Name: "user"})
	_, err := builder.Build()
	assert.EqualError(t, err, "user: missing config type")
}

func TestBuildSchema(t *testing.T) {
	actual := Schema().
		Realm(
			Realm("myrealm").
				Module(
					Module("service").
						Decl(&schema.Config{Name: "user", Type: &schema.Int{}}).
						MustBuild()).
				MustBuild()).
		MustBuild()
	expected := &schema.Schema{
		Realms: []*schema.Realm{
			{
				Name: "myrealm",
				Modules: []*schema.Module{
					schema.Builtins(),
					{
						Name: "service",
						Decls: []schema.Decl{
							&schema.Config{
								Name: "user",
								Type: &schema.Int{},
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, actual)
}
