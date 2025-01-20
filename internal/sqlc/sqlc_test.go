package sqlc

import (
	"context"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/scaffolder"
)

func TestAddQueriesToSchema(t *testing.T) {
	t.Skip("flaky")
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
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Debug}))
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
						Query:   "SELECT data FROM requests",
						Command: "many",
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
						Query:   "INSERT INTO requests (data) VALUES (?)",
						Command: "exec",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}
