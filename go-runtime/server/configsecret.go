package server

import (
	"reflect"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
)

func Config[T ftl.ConfigType](module, name string) reflection.VerbResource {
	cfg := ftl.Config[T]{Ref: reflection.Ref{Module: module, Name: name}}
	return func() reflect.Value {
		return reflect.ValueOf(cfg)
	}
}

func Secret[T ftl.SecretType](module, name string) reflection.VerbResource {
	secret := ftl.Secret[T]{Ref: reflection.Ref{Module: module, Name: name}}
	return func() reflect.Value {
		return reflect.ValueOf(secret)
	}
}
