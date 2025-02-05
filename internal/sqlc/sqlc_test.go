//go:build integration

package sqlc

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/scaffolder"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestAddQueriesToSchema(t *testing.T) {
	tmpDir := t.TempDir()

	err := scaffolder.Scaffold("testdata", tmpDir, nil)
	assert.NoError(t, err)
	mc := moduleconfig.ModuleConfig{
		Dir:                   tmpDir,
		Module:                "test",
		SQLMigrationDirectory: "db/schema",
		SQLQueryDirectory:     "db/queries",
		DeployDir:             ".ftl",
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	out := &schema.Schema{}
	updated, err := AddQueriesToSchema(ctx, tmpDir, mc.Abs(), out)
	assert.NoError(t, err)
	assert.True(t, updated)

	var actual *schema.Module
	for _, d := range out.Modules {
		if d.Name == "test" {
			actual = d
			break
		}
	}
	schema.SortModuleDecls(actual)

	expected := &schema.Module{
		Name: "test",
		Decls: []schema.Decl{
			&schema.Data{
				Name: "CreateRequestQuery",
				Fields: []*schema.Field{
					{
						Name: "data",
						Type: &schema.String{},
						Metadata: []schema.Metadata{
							&schema.MetadataSQLColumn{
								Table: "requests",
								Name:  "data",
							},
						},
					},
				},
			},
			&schema.Data{
				Name: "GetRequestDataResult",
				Fields: []*schema.Field{
					{
						Name: "data",
						Type: &schema.String{},
						Metadata: []schema.Metadata{
							&schema.MetadataSQLColumn{
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
						Query:   "INSERT INTO requests (data) VALUES (?)",
						Command: "exec",
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
						Query:   "SELECT data FROM requests",
						Command: "many",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}
