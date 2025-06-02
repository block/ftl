package change

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
)

func TestBreaking(t *testing.T) {
	tests := []struct {
		name       string
		prev, next string
		breaking   string
	}{
		{
			name:     "TypeChange",
			prev:     "String",
			next:     "Int",
			breaking: "previous.schema:1:1: type changed from string to int",
		},
		{
			name: "NoTypeChange",
			prev: "String",
			next: "String",
		},
		{
			name:     "DataRename",
			prev:     `data Request {}`,
			next:     `data Response {}`,
			breaking: `previous.schema:1:1: Request: data name changed from Request to Response`,
		},
		{
			name: "FieldTypeChange",
			prev: `data Request {
				name String
				age Int
			}`,
			next: `data Request {
				name String
				age Time
			}`,
			breaking: `previous.schema:3:9: Request: age: type changed from int to time`,
		},
		{
			name: "FieldRemoved",
			prev: `data Request {
				name String
				age Int
			}`,
			next: `data Request {
				name String
				born Time
			}`,
			breaking: `previous.schema:3:5: Request: field age removed`,
		},
		{
			name:     "ArrayElementTypeChange",
			prev:     `[String]`,
			next:     `[Int]`,
			breaking: `previous.schema:1:2: array: type changed from string to int`,
		},
		{
			name:     "TypeAliasRenamed",
			prev:     `typealias A Int`,
			next:     `typealias B Int`,
			breaking: `previous.schema:1:1: typealias A: type alias name changed to B`,
		},
		{
			name:     "TypeAliasChanged",
			prev:     `typealias A Int`,
			next:     `typealias A String`,
			breaking: `previous.schema:1:13: typealias A: type changed from int to string`,
		},
		{
			name: "EnumOrderChanged",
			prev: `enum Enum: Int {
				A = 0
				B = 1
			}`,
			next: `enum Enum: Int {
				B = 1
				A = 0
			}`,
		},
		{
			name: "EnumValueAdded",
			prev: `enum Enum: Int {
				A = 0
				B = 1
			}`,
			next: `enum Enum: Int {
				A = 0
				B = 1
				C = 2
			}`,
			breaking: `previous.schema:1:1: enum values changed`,
		},
		{
			name: "EnumValueRemoved",
			prev: `enum Enum: Int {
				A = 0
				B = 1
			}`,
			next: `enum Enum: Int {
				A = 0
			}`,
			breaking: `previous.schema:1:1: enum values changed`,
		},
		{
			name: "TypeEnumOrder",
			prev: `enum Enum {
				Age Int
				Name String
			}`,
			next: `enum Enum {
				Name String
				Age Int
			}`,
		},
		{
			name: "TypeEnumTypeChange",
			prev: `enum Enum {
				Age Int
				Name String
			}`,
			next: `enum Enum {
				Age Time
				Name String
			}`,
			breaking: `previous.schema:1:1: enum values changed`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, err := schema.ParseType("previous.schema", tt.prev)
			assert.NoError(t, err)
			next, err := schema.ParseType("previous.schema", tt.next)
			assert.NoError(t, err)
			err = Breaking(None[*schema.Schema](), prev, next)
			if tt.breaking != "" {
				assert.EqualError(t, err, tt.breaking)
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
