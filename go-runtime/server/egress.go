package server

import (
	"reflect"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
)

func Egress(name string) reflection.VerbResource {
	cfg := ftl.EgressTarget{Name: name}
	return func() reflect.Value {
		return reflect.ValueOf(cfg)
	}
}
