package sqlc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/scaffolder"
)

func TestGenerate(t *testing.T) {
	if err := os.RemoveAll(filepath.Join(os.TempDir(), ".ftl")); err != nil {
		t.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "sqlc-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	schemaDir := filepath.Join(tmpDir, "schema")
	err = scaffolder.Scaffold("testdata", schemaDir, nil)
	assert.NoError(t, err)
	pc := projectconfig.Config{
		Path: tmpDir,
	}
	mc := moduleconfig.ModuleConfig{
		Dir:                   tmpDir,
		Module:                "test",
		SQLMigrationDirectory: "schema",
		DeployDir:             ".ftl",
	}

	actual, err := Generate(pc, mc)
	assert.NoError(t, err)

	expected := &schema.Module{
		Name: "test",
		Decls: []schema.Decl{
			&schema.Data{
				Name: "GetRequestDataResult",
				Fields: []*schema.Field{
					{
						Name: "data",
						Type: &schema.String{},
						Metadata: []schema.Metadata{
							&schema.MetadataDBColumn{
								Table: "requests",
								Name:  "data",
							},
						},
					},
				},
			},
			&schema.Verb{
				Name:    "GetRequestData",
				Request: &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{
					Module: "test",
					Name:   "GetRequestDataResult",
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataSQLQuery{
						Query: "SELECT data FROM requests",
					},
				},
			},
			&schema.Data{
				Name: "CreateRequestQuery",
				Fields: []*schema.Field{
					{
						Name: "data",
						Type: &schema.String{},
						Metadata: []schema.Metadata{
							&schema.MetadataDBColumn{
								Table: "requests",
								Name:  "data",
							},
						},
					},
				},
			},
			&schema.Verb{
				Name:     "CreateRequest",
				Request:  &schema.Ref{Module: "test", Name: "CreateRequestQuery"},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataSQLQuery{
						Query: "INSERT INTO requests (data) VALUES (?)",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}
