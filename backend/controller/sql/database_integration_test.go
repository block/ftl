//go:build integration

package sql_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestPostgres(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java", "kotlin"),
		// deploy real module against "testdb"
		in.CopyModule("database"),
		in.Deploy("database"),
		in.Call[in.Obj, in.Obj]("database", "insert", in.Obj{"data": "hello"}, nil),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		in.IfLanguage("go", in.ExecModuleTest("database")),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestMySQL(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java", "kotlin"),
		// deploy real module against "testdb"
		in.CopyModule("mysql"),
		in.Deploy("mysql"),
		in.Call[in.Obj, in.Obj]("mysql", "insert", in.Obj{"data": "hello"}, nil),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
		in.IfLanguage("go", in.ExecModuleTest("mysql")),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
	)
}

func TestSQLVerbs(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java", "kotlin"),
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
		in.Call[int, in.Obj]("mysql", "getTestType", 1, func(t testing.TB, response in.Obj) {
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

func TestTransactions(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java", "kotlin"),
		in.CopyModule("database"),
		in.Deploy("database"),

		// successful transaction
		in.Call[in.Obj, in.Obj]("database", "transactionInsert", in.Obj{
			"items": []string{"item1", "item2", "item3"},
		}, func(t testing.TB, response in.Obj) {
			count := response["count"].(float64)
			assert.Equal(t, float64(3), count, "Transaction should have inserted 3 items")
		}),
		in.QueryRow("database_testdb", "SELECT data FROM requests WHERE data = 'item1'", "item1"),
		in.QueryRow("database_testdb", "SELECT data FROM requests WHERE data = 'item2'", "item2"),
		in.QueryRow("database_testdb", "SELECT data FROM requests WHERE data = 'item3'", "item3"),
		in.QueryRow("database_testdb", "SELECT COUNT(*) FROM requests", float64(3)),

		// rollback
		in.ExpectError(
			in.Call[in.Obj, in.Obj]("database", "transactionRollback", in.Obj{
				"items": []string{"should-not-be-committed"},
			}, nil),
			"deliberate error to test rollback",
		),
		in.QueryRow("database_testdb", "SELECT COUNT(*) FROM requests WHERE data = 'should-not-be-committed'", float64(0)),
		in.QueryRow("database_testdb", "SELECT COUNT(*) FROM requests", float64(3)),
	)
}
