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

func Database[T any](dbName string, init func(ref Ref) (ReflectedDatabase, *once.Handle[*sql.DB])) Registree {
	ref := Ref{
		Module: moduleForType(reflect.TypeFor[T]()),
		Name:   dbName,
	}
	return func(t *TypeRegistry) {
		db, handle := init(ref)
		t.dbHandles[reflect.TypeFor[T]()] = &ReflectedDatabaseHandle{
			ReflectedDatabase: db,
			DB:                handle,
		}
		t.databases[ref] = db
	}
}

func Transaction(verb any, dbName string) Registree {
	ref := FuncRef(verb)
	return func(t *TypeRegistry) {
		t.transactionVerbs[ref] = Ref{Module: ref.Module, Name: dbName}
	}
}
