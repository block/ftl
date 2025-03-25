package reflection

import (
	"reflect"
)

type CommandType int

const (
	CommandTypeExec CommandType = iota
	CommandTypeOne
	CommandTypeMany
)

func Query(
	module string,
	verbName string,
	queryFunc any,
) Registree {
	ref := Ref{
		Module: module,
		Name:   verbName,
	}
	return func(t *TypeRegistry) {
		vi := verbCall{
			ref:  ref,
			args: []reflect.Value{},
			fn:   reflect.ValueOf(queryFunc),
		}
		t.verbCalls[ref] = vi
	}
}
