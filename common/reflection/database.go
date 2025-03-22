package reflection

import (
	"database/sql"
	"reflect"

	"github.com/alecthomas/types/once"
)

type ReflectedDatabase struct {
	DBType string
	// configs
	Name string
}
type ReflectedDatabaseHandle struct {
	ReflectedDatabase
	DB *once.Handle[*sql.DB]
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
