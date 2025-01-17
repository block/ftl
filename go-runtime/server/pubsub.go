package server

import (
	"reflect"

	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/ftl/reflection"
	"github.com/block/ftl/internal/schema"
)

func TopicHandle[E any](module, name string) reflection.VerbResource {
	handle := ftl.TopicHandle[E]{Ref: &schema.Ref{
		Name:   name,
		Module: module,
	}}
	return func() reflect.Value {
		return reflect.ValueOf(handle)
	}
}
