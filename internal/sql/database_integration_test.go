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
		in.CopyModule("postgres"),
		in.Deploy("postgres"),
		in.Call[in.Obj, in.Obj]("postgres", "insert", in.Obj{"data": "hello"}, nil),
		in.QueryRow("postgres_testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		// TODO(worstell): uncomment once we refactor unit test support
		// in.IfLanguage("go", in.ExecModuleTest("postgres")),
		in.QueryRow("postgres_testdb", "SELECT data FROM requests", "hello"),

		// TODO(worstell): Make slices work in Postgres
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
		// TODO(worstell): uncomment once we refactor unit test support
		// in.IfLanguage("go", in.ExecModuleTest("mysql")),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),

		// Add more data for SLICE testing
		in.Call[in.Obj, in.Obj]("mysql", "insert", in.Obj{"data": "apple"}, nil),
		in.Call[in.Obj, in.Obj]("mysql", "insert", in.Obj{"data": "banana"}, nil),

		// Test SLICE pattern with IN clause
		in.Call[[]string, []in.Obj]("mysql", "findMultiple", []string{"hello", "apple", "banana"}, func(t testing.TB, response []in.Obj) {
			assert.Equal(t, 3, len(response), "Should find all 3 items")
		}),

		// Test multiple SLICE patterns in one query
		in.Call[in.Obj, []in.Obj]("mysql", "findByDataAndIds", in.Obj{
			"dataValues": []string{"hello", "apple"},
			"ids":        []int{1},
		}, func(t testing.TB, response []in.Obj) {
			assert.True(t, len(response) == 1, "Should find only items matching both criteria")
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
		in.CopyModule("postgres"),
		in.Deploy("postgres"),

		// successful transaction
		in.Call("postgres", "transactionInsert", in.Obj{
			"items": []string{"item1", "item2", "item3"},
		}, func(t testing.TB, response in.Obj) {
			count := response["count"].(float64)
			assert.Equal(t, float64(3), count, "Transaction should have inserted 3 items")
		}),
		in.QueryRow("postgres_testdb", "SELECT data FROM requests WHERE data = 'item1'", "item1"),
		in.QueryRow("postgres_testdb", "SELECT data FROM requests WHERE data = 'item2'", "item2"),
		in.QueryRow("postgres_testdb", "SELECT data FROM requests WHERE data = 'item3'", "item3"),
		in.QueryRow("postgres_testdb", "SELECT COUNT(*) FROM requests", float64(3)),

		// rollback
		in.ExpectError(
			in.Call[in.Obj, in.Obj]("postgres", "transactionRollback", in.Obj{
				"items": []string{"should-not-be-committed"},
			}, nil),
			"deliberate error to test rollback",
		),
		in.QueryRow("postgres_testdb", "SELECT COUNT(*) FROM requests WHERE data = 'should-not-be-committed'", float64(0)),
		in.QueryRow("postgres_testdb", "SELECT COUNT(*) FROM requests", float64(3)),
	)
}

func TestLists(t *testing.T) {
	in.Run(t,
		in.WithDebugLogging(),
		in.WithLanguages("kotlin"),
		in.CopyModule("mysql"),
		in.Deploy("mysql"),

		in.Call[in.Obj, in.Obj]("mysql", "insertTestData", in.Obj{"intVal": 1, "floatVal": 2, "textVal": "test"}, nil),
		in.Call[in.Obj, []in.Obj]("mysql", "querySlices", in.Obj{"ints": []int{1}, "floats": []float64{2.0}, "texts": []string{"test"}}, func(t testing.TB, response []in.Obj) {
			assert.Equal(t, 1, len(response))
			assert.Equal(t, 1, response[0]["intVal"])
			assert.Equal(t, 2, response[0]["floatVal"])
			assert.Equal(t, "test", response[0]["textVal"])
		}),
	)
}
