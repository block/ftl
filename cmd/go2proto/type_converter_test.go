package main

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/alecthomas/assert/v2"
)

type Test[T any] struct {
	Value T
}

func TestConvert(t *testing.T) {
	converter := &Converter{
		Rules: []Conversion{
			BasicToPtr{},
			PtrToBasic{},
		},
	}

	tests := []struct {
		from types.Type
		to   types.Type
		expr Expr
	}{
		{
			from: types.Typ[types.Int],
			to:   types.NewPointer(types.Typ[types.Int]),
			expr: "&a",
		},
		{
			from: types.NewPointer(types.Typ[types.Float32]),
			to:   types.Typ[types.Float32],
			expr: "*a",
		},
	}

	typeParam := types.NewTypeParam(types.NewTypeName(0, nil, "T", nil), nil)
	field := types.NewVar(0, nil, "value", typeParam)
	structType := types.NewStruct([]*types.Var{field}, nil)
	obj := types.NewTypeName(0, nil, "Foo", structType)
	namedType := types.NewNamed(obj, structType, nil)

	fmt.Printf("%+v\n", namedType.TypeParams())

	for _, test := range tests {
		expr, err := converter.Convert(ConversionState{Type: test.from, Expr: "a"}, test.to)
		assert.NoError(t, err)
		assert.Equal(t, test.expr, expr)
	}
}
