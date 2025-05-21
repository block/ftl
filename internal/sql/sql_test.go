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
	"github.com/block/ftl/internal/projectconfig"
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
	}.FillDefaultsAndValidate(moduleconfig.CustomDefaults{}, projectconfig.Config{Name: "test"})
	assert.NoError(t, err)

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	out := &schema.Schema{Realms: []*schema.Realm{{Name: "test"}}}
	err = AddDatabaseDeclsToSchema(ctx, tmpDir, mc.Abs(), out)
	assert.NoError(t, err)

	var actual *schema.Module
	for _, d := range out.InternalModules() {
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
				Name: "AuthorRow",
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     10,
				},
				Fields: []*schema.Field{{
					Name: "id",
					Type: &schema.Int{},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "id", Table: "authors"},
					},
				}, {
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "birthYear",
					Type: &schema.Optional{Type: &schema.Int{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "birth_year", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Name: "GetAuthorInfoMySqlRow",
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     13,
				},
				Fields: []*schema.Field{{
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Name: "GetAuthorInfoPsqlRow",
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     13,
				},
				Fields: []*schema.Field{{
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Name: "GetManyAuthorsInfoMySqlRow",
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     16,
				},
				Fields: []*schema.Field{{
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Name: "GetManyAuthorsInfoPsqlRow",
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     16,
				},
				Fields: []*schema.Field{{
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Data{
				Name: "PsqldbAuthorRow",
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     10,
				},
				Fields: []*schema.Field{{
					Name: "id",
					Type: &schema.Int{},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "id", Table: "authors"},
					},
				}, {
					Name: "bio",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "bio", Table: "authors"},
					},
				}, {
					Name: "birthYear",
					Type: &schema.Optional{Type: &schema.Int{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "birth_year", Table: "authors"},
					},
				}, {
					Name: "hometown",
					Type: &schema.Optional{Type: &schema.String{}},
					Metadata: []schema.Metadata{
						&schema.MetadataSQLColumn{Name: "hometown", Table: "authors"},
					},
				}},
				Metadata: []schema.Metadata{
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     4,
				},
				Name:     "createRequestMySql",
				Request:  &schema.String{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
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
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     4,
				},
				Name:     "createRequestPsql",
				Request:  &schema.String{},
				Response: &schema.Unit{},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
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
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     7,
				},
				Name:     "getAllAuthorsMySql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{Module: "test", Name: "AuthorRow"}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT id, bio, birth_year, hometown FROM authors",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     7,
				},
				Name:     "getAllAuthorsPsql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{Module: "test", Name: "PsqldbAuthorRow"}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT id, bio, birth_year, hometown FROM authors",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     10,
				},
				Name:     "getAuthorByIdMySql",
				Request:  &schema.Int{},
				Response: &schema.Ref{Module: "test", Name: "AuthorRow"},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT id, bio, birth_year, hometown FROM authors WHERE id = ?",
						Command: "one",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     10,
				},
				Name:     "getAuthorByIdPsql",
				Request:  &schema.Int{},
				Response: &schema.Ref{Module: "test", Name: "PsqldbAuthorRow"},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT id, bio, birth_year, hometown FROM authors WHERE id = $1",
						Command: "one",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     13,
				},
				Name:     "getAuthorInfoMySql",
				Request:  &schema.Int{},
				Response: &schema.Ref{Module: "test", Name: "GetAuthorInfoMySqlRow"},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT bio, hometown FROM authors WHERE id = ?",
						Command: "one",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     13,
				},
				Name:     "getAuthorInfoPsql",
				Request:  &schema.Int{},
				Response: &schema.Ref{Module: "test", Name: "GetAuthorInfoPsqlRow"},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT bio, hometown FROM authors WHERE id = $1",
						Command: "one",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     16,
				},
				Name:     "getManyAuthorsInfoMySql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{Module: "test", Name: "GetManyAuthorsInfoMySqlRow"}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT bio, hometown FROM authors",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     16,
				},
				Name:     "getManyAuthorsInfoPsql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.Ref{Module: "test", Name: "GetManyAuthorsInfoPsqlRow"}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
					},
					&schema.MetadataSQLQuery{
						Query:   "SELECT bio, hometown FROM authors",
						Command: "many",
					},
					&schema.MetadataGenerated{},
				},
			},
			&schema.Verb{
				Pos: schema.Position{
					Filename: "db/mysql/mysqldb/queries/queries.sql",
					Line:     1,
				},
				Name:     "getRequestDataMySql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.String{}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "mysqldb"}},
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
					Filename: "db/postgres/psqldb/queries/queries.sql",
					Line:     1,
				},
				Name:     "getRequestDataPsql",
				Request:  &schema.Unit{},
				Response: &schema.Array{Element: &schema.String{}},
				Metadata: []schema.Metadata{
					&schema.MetadataDatabases{
						Uses: []*schema.Ref{{Module: "test", Name: "psqldb"}},
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
