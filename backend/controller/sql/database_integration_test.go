//go:build integration

package sql_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	in "github.com/block/ftl/internal/integration"
)

func TestDatabase(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		// deploy real module against "testdb"
		in.CopyModule("database"),
		in.Deploy("database"),
		in.Call[in.Obj, in.Obj]("database", "insert", in.Obj{"data": "hello", "id": 1}, nil),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		in.IfLanguage("go", in.ExecModuleTest("database")),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestMySQL(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		// deploy real module against "testdb"
		in.CopyModule("mysql"),
		in.Deploy("mysql"),
		in.Call[in.Obj, in.Obj]("mysql", "insert", in.Obj{"data": "hello"}, nil),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
		// TODO: handle query verbs in ftltest
		// in.IfLanguage("go", in.ExecModuleTest("mysql")),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
		in.IfLanguage("go", in.VerifySchemaVerb("mysql", "createRequest", func(ctx context.Context, t testing.TB, sch *schemapb.Schema, verb *schemapb.Verb) {
			assert.True(t, verb.Response.GetUnit() != nil, "response was not a unit")
			assert.True(t, verb.Request.GetRef() != nil, "request was not a ref")
			fullSchema, err := schema.FromProto(sch)
			assert.NoError(t, err, "failed to convert schema")
			req := fullSchema.Resolve(must.Get(schema.RefFromProto(verb.Request.GetRef())))
			assert.True(t, req.Ok(), "request not found")

			if data, ok := req.MustGet().(*schema.Data); ok {
				assert.Equal(t, "CreateRequestQuery", data.Name)
				assert.Equal(t, 1, len(data.Fields))
				assert.Equal(t, "data", data.Fields[0].Name)
			} else {
				assert.False(t, true, "request not data")
			}
		})),
		in.IfLanguage("go", in.VerifySchemaVerb("mysql", "getRequestData", func(ctx context.Context, t testing.TB, sch *schemapb.Schema, verb *schemapb.Verb) {
			assert.True(t, verb.Response.GetArray() != nil, "response was not an array")
			assert.True(t, verb.Response.GetArray().Element.GetRef() != nil, "array element was not a ref")
			assert.True(t, verb.Request.GetUnit() != nil, "request was not a unit")
			fullSchema, err := schema.FromProto(sch)
			assert.NoError(t, err, "failed to convert schema")

			resp := fullSchema.Resolve(must.Get(schema.RefFromProto(verb.Response.GetArray().Element.GetRef())))
			assert.True(t, resp.Ok(), "response not found")

			if data, ok := resp.MustGet().(*schema.Data); ok {
				assert.Equal(t, "GetRequestDataResult", data.Name)
				assert.Equal(t, 1, len(data.Fields))
				assert.Equal(t, "data", data.Fields[0].Name)
			} else {
				assert.False(t, true, "response not data")
			}
		})),
	)
}

func TestSQLVerbs(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go"),
		in.CopyModule("mysql"),
		in.Deploy("mysql"),

		// Test EXEC operation - insert a record with all types
		in.Call[in.Obj, in.Obj]("mysql", "insertTestTypes", in.Obj{
			"intVal":      42,
			"floatVal":    3.14,
			"textVal":     "hello world",
			"boolVal":     true,
			"timeVal":     "2024-01-01T12:00:00Z",
			"optionalVal": "optional value",
		}, nil),

		// Test ONE operation - get the inserted record
		in.Call[in.Obj, in.Obj]("mysql", "getTestType", in.Obj{"id": 1}, func(t testing.TB, response in.Obj) {
			intVal := response["intVal"].(float64)
			floatVal := response["floatVal"].(float64)

			assert.Equal(t, float64(42), intVal)
			assert.Equal(t, 3.14, floatVal)
			assert.Equal(t, "hello world", response["textVal"])
			assert.Equal(t, true, response["boolVal"])
			assert.Equal(t, "2024-01-01T12:00:00Z", response["timeVal"])
			// todo: make optionals work with test helper
			// assert.Equal(t, "optional value", response["optionalVal"])
		}),

		// Test MANY operation - get all records
		in.Call[in.Obj, []in.Obj]("mysql", "getAllTestTypes", in.Obj{}, func(t testing.TB, response []in.Obj) {
			record := response[0]
			intVal := record["intVal"].(float64)
			floatVal := record["floatVal"].(float64)

			assert.Equal(t, float64(42), intVal)
			assert.Equal(t, 3.14, floatVal)
			assert.Equal(t, "hello world", record["textVal"])
			assert.Equal(t, true, record["boolVal"])
			assert.Equal(t, "2024-01-01T12:00:00Z", record["timeVal"])
			// todo: make optionals work with test helper
			// assert.Equal(t, "optional value", record["optionalVal"])
		}),
	)
}
