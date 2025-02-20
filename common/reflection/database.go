package reflection

import (
	"database/sql"
	"reflect"

	"github.com/alecthomas/types/once"
)

type ReflectedDatabaseHandle struct {
	DBType string
	DB     *once.Handle[*sql.DB]

	// configs
	Name string
}

func Database[T any](dbname string, init func(ref Ref) *ReflectedDatabaseHandle) Registree {
	ref := Ref{
		Module: moduleForType(reflect.TypeFor[T]()),
		Name:   dbname,
	}
	return func(t *TypeRegistry) {
		t.databases[reflect.TypeFor[T]()] = init(ref)
	}
}
