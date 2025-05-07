package transform

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
		optionalBytesType  = must.Get(types.Instantiate(tctx, optionalType, []types.Type{bytesType}, false))
	)
	tests := []struct {
		name     string
		input    types.Type
		output   types.Type
		expected string
		ok       bool
	}{
		{"StringToBytes",
			stringType,
			bytesType,
			`result.Map(result.Ok(input), stringToBytes)`,
			true},
		{"MarshalBinary",
			urlType,
			bytesType,
			`result.Map(result.Ok(input), marshalBinary)`,
			true},
		{"ToOptional",
			optionalStringType,
			stringType,
			"result.Map(result.Ok(input), unwrapOptional)",
			true},
		{"OptionalStringToBytes",
			optionalStringType,
			bytesType,
			"result.Map(result.Map(result.Ok(input), unwrapOptional), stringToBytes)",
			true},
		{"OptionalBytesToString",
			optionalBytesType,
			stringType,
			"result.Map(result.Map(result.Ok(input), unwrapOptional), bytesToString)",
			true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := Var("input", test.input)
			out, ok := Transform(input, test.output)
			assert.Equal(t, test.ok, ok, "could not find transform from %s to %s", test.input, test.output)
			if ok {
				assert.Equal(t, test.expected, out.Expr())
			}
		})
	}
}
