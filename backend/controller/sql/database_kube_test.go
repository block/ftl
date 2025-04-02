//go:build infrastructure

package sql_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestSQLVerbsInKube(t *testing.T) {
	in.Run(t,
		in.WithKubernetes("--set", "ftl.provisioner.modulePerNamespace=true"),
		in.WithLanguages("kotlin"),
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
