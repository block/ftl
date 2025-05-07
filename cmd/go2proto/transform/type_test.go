package transform

import (
	"fmt"
	"go/types"
	"testing"
)

func TestType(t *testing.T) {
	t.Run("GoTypes", func(t *testing.T) {
		obj := loadObject("github.com/block/ftl/cmd/go2proto/testdata", "Message").(*types.TypeName).Type()
		fmt.Printf("%T\n", obj)
		testType(t, FromGoTypes(obj))
	})
}

func testType(t *testing.T, typ Type) {
	for _, method := range typ.Methods() {
		fmt.Println(method.Name)
	}
	for _, field := range typ.Fields() {
		fmt.Println("field", field.Name, field.Type)
	}
}
