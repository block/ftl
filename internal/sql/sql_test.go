//go:build integration

package sql

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/scaffolder"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

func TestAddDatabaseDeclsToSchema(t *testing.T) {
	tmpDir := t.TempDir()

	err := scaffolder.Scaffold(filepath.Join("testdata", "database"), tmpDir, nil)
	assert.NoError(t, err)
	mc, err := moduleconfig.UnvalidatedModuleConfig{
		Dir:        tmpDir,
		Module:     "test",
		SQLRootDir: "db",
		DeployDir:  ".ftl",
	}.FillDefaultsAndValidate(moduleconfig.CustomDefaults{})
	assert.NoError(t, err)

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	out := &schema.Schema{}
	err = AddDatabaseDeclsToSchema(ctx, tmpDir, mc.Abs(), out)
	assert.NoError(t, err)

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
			&schema.Database{Name: "mysqldb", Type: schema.MySQLDatabaseType},
			&schema.Database{Name: "psqldb", Type: schema.PostgresDatabaseType},
			&schema.Data{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/mysql/mysqldb/queries/queries.sql"),
					Line:     4,
				},
				Name: "CreateRequestMySqlQuery",
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
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/postgres/psqldb/queries/queries.sql"),
					Line:     4,
				},
				Name: "CreateRequestPsqlQuery",
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
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/mysql/mysqldb/queries/queries.sql"),
					Line:     1,
				},
				Name: "GetRequestDataMySqlResult",
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
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/postgres/psqldb/queries/queries.sql"),
					Line:     1,
				},
				Name: "GetRequestDataPsqlResult",
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
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/mysql/mysqldb/queries/queries.sql"),
					Line:     4,
				},
				Name:     "createRequestMySql",
				Request:  &schema.Ref{Module: "test", Name: "CreateRequestMySqlQuery"},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{
							{
								Module: "test",
								Name:   "mysqldb",
							},
						},
					},
					&schema.MetadataSQLQuery{
						Query:   "INSERT INTO requests (data) VALUES (?)",
						Command: "exec",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/postgres/psqldb/queries/queries.sql"),
					Line:     4,
				},
				Name:     "createRequestPsql",
				Request:  &schema.Ref{Module: "test", Name: "CreateRequestPsqlQuery"},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{
							{
								Module: "test",
								Name:   "psqldb",
							},
						},
					},
					&schema.MetadataSQLQuery{
						Query:   "INSERT INTO requests (data) VALUES ($1)",
						Command: "exec",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/mysql/mysqldb/queries/queries.sql"),
					Line:     1,
				},
				Name:    "getRequestDataMySql",
				Request: &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{
					Module: "test",
					Name:   "GetRequestDataMySqlResult",
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{
							{
								Module: "test",
								Name:   "mysqldb",
							},
						},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT data FROM requests",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: filepath.Join(tmpDir, "db/postgres/psqldb/queries/queries.sql"),
					Line:     1,
				},
				Name:    "getRequestDataPsql",
				Request: &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{
					Module: "test",
					Name:   "GetRequestDataPsqlResult",
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{
							{
								Module: "test",
								Name:   "psqldb",
							},
						},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT data FROM requests",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
		},
	}

	assert.Equal(t, expected, actual, "expected: %s\nactual: %s", expected, actual)
}
