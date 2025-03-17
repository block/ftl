package schema

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTransformFromAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			enum TypeEnum {
				A test.Inner
				B String
			}
			
			data Inner {
				waz String +alias json "foo"
			}

			data Test {
				scalar String +alias json "bar"
				inner test.Inner
				array [test.Inner]
				map {String: test.Inner}
				optional test.Inner
				typeEnum test.TypeEnum
			}
		}
		`

	sch, err := ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := TransformFromAliasedFields(&Ref{Module: "test", Name: "Test"}, sch, map[string]any{
		"bar": "value",
		"inner": map[string]any{
			"foo": "value",
		},
		"array": []any{
			map[string]any{
				"foo": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"foo": "value",
			},
		},
		"optional": map[string]any{
			"foo": "value",
		},
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"foo": "value"},
		},
	})
	expected := map[string]any{
		"scalar": "value",
		"inner": map[string]any{
			"waz": "value",
		},
		"array": []any{
			map[string]any{
				"waz": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"waz": "value",
			},
		},
		"optional": map[string]any{
			"waz": "value",
		},
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"waz": "value"},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestTransformToAliasedFields(t *testing.T) {
	schemaText := `
		module test {
			enum TypeEnum {
				A test.Inner
				B String
			}

			data Inner {
				waz String +alias json "foo"
			}

			data Test {
				scalar String +alias json "bar"
				inner test.Inner
				array [test.Inner]
				map {String: test.Inner}
				optional test.Inner
				typeEnum test.TypeEnum
			}
		}
		`

	sch, err := ParseString("test", schemaText)
	assert.NoError(t, err)
	actual, err := TransformToAliasedFields(&Ref{Module: "test", Name: "Test"}, sch, map[string]any{
		"scalar": "value",
		"inner": map[string]any{
			"waz": "value",
		},
		"array": []any{
			map[string]any{
				"waz": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"waz": "value",
			},
		},
		"optional": map[string]any{
			"waz": "value",
		},
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"waz": "value"},
		},
	})
	expected := map[string]any{
		"bar": "value",
		"inner": map[string]any{
			"foo": "value",
		},
		"array": []any{
			map[string]any{
				"foo": "value",
			},
		},
		"map": map[string]any{
			"key": map[string]any{
				"foo": "value",
			},
		},
		"optional": map[string]any{
			"foo": "value",
		},
		"typeEnum": map[string]any{
			"name":  "A",
			"value": map[string]any{"foo": "value"},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestValidateJSONCall(t *testing.T) {
	schemaText := `
module echo {
  export data EchoRequest {
    name String? +alias json "name"
	age Int? +alias json "age"
	weight Float? +alias json "weight"
  }

  export data EchoResponse {
    message String +alias json "message"
  }

  export verb echo(echo.EchoRequest) echo.EchoResponse
}`

	sch, err := ParseString("test", schemaText)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		ref     *Ref
		input   string
		wantErr string
	}{
		{
			name:  "valid input",
			ref:   &Ref{Module: "echo", Name: "echo"},
			input: `{"name": "juho", "age": 123, "weight": 0.045}`,
		},
		{
			name:  "valid integer as float",
			ref:   &Ref{Module: "echo", Name: "echo"},
			input: `{"name": "juho", "age": 123.0, "weight": 0.045}`,
		},
		{
			name:    "invalid float for integer",
			ref:     &Ref{Module: "echo", Name: "echo"},
			input:   `{"name": "juho", "age": 123.5, "weight": 0.045}`,
			wantErr: "age has wrong type, expected Int found float64",
		},
		{
			name:    "invalid input",
			ref:     &Ref{Module: "echo", Name: "echo"},
			input:   `{"name": "juho", "age": 123, "weight": "too much"}`,
			wantErr: "weight has wrong type, expected Float found string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input map[string]any
			err := json.Unmarshal([]byte(tt.input), &input)
			assert.NoError(t, err)

			err = ValidateJSONCall([]byte(tt.input), tt.ref, sch)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
