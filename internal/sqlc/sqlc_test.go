package sqlc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/scaffolder"
)

func TestAddQueriesToSchema(t *testing.T) {
	if err := os.RemoveAll(filepath.Join(os.TempDir(), ".ftl")); err != nil {
		t.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "sqlc-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = scaffolder.Scaffold("testdata", tmpDir, nil)
	assert.NoError(t, err)
	mc := moduleconfig.ModuleConfig{
		Dir:                   tmpDir,
		Module:                "test",
		SQLMigrationDirectory: "db/schema",
		SQLQueryDirectory:     "db/queries",
		DeployDir:             ".ftl",
	}

	out := &schema.Schema{}
	updated, err := AddQueriesToSchema(tmpDir, mc.Abs(), out)
	assert.NoError(t, err)
	assert.True(t, updated)

	var actual *schema.Module
	for _, d := range out.Modules {
		if d.Name == "test" {
			actual = d
			break
		}
	}

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
				Name:    "getRequestData",
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
				Name:     "createRequest",
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
