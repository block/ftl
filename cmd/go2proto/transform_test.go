package main

import (
	"go/types"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
)

func TestTransform(t *testing.T) {
	var (
		tctx               = types.NewContext()
		stringType         = basicType("string")
		bytesType          = types.NewSlice(basicType("byte"))
		urlType            = loadObject("net/url", "URL").(*types.TypeName).Type()
		optionalType       = loadObject("github.com/alecthomas/types/optional", "Option").(*types.TypeName).Type()
		optionalStringType = must.Get(types.Instantiate(tctx, optionalType, []types.Type{stringType}, false))
		// optionalBytesType  = must.Get(types.Instantiate(tctx, optionalType, []types.Type{bytesType}, false))
	)
	tests := []struct {
		name     string
		input    types.Type
		output   types.Type
		expected string
	}{
		{"StringToBytes",
			stringType,
			bytesType,
			`result.Map(result.Ok(hello), stringToBytes)`,
		},
		{"MarshalBinary",
			urlType,
			bytesType,
			`result.Map(result.Ok(hello), marshalBinary)`},
		{"ToOptional",
			optionalStringType,
			stringType,
			"result.Map(result.Ok(hello), unwrapOptional)",
		},
		{"OptionalStringToBytes",
			optionalStringType,
			bytesType,
			"result.Map(result.Map(result.Ok(hello), unwrapOptional), stringToBytes)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := Var("hello", test.input)
			out, ok := Transform(input, test.output)
			assert.True(t, ok, "could not find transform")
			assert.Equal(t, test.expected, out.Expr())
		})
	}
}
