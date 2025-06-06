package ftl

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/go-runtime/internal"
)

type MapHandle[T, U any] struct {
	fn     func(context.Context, T) (U, error)
	handle Handle[T]
}

// Get the mapped value.
func (mh *MapHandle[T, U]) Get(ctx context.Context) U {
	value := mh.handle.Get(ctx)
	out := internal.FromContext(ctx).CallMap(ctx, mh, value, func(ctx context.Context) (any, error) {
		return errors.WithStack2(mh.fn(ctx, value))
	})
	u, ok := out.(U)
	if !ok {
		panic(fmt.Sprintf("output object %v is not compatible with expected type %T", out, *new(U)))
	}
	return u
}

// Map an FTL resource type to a new type.
func Map[T, U any](getter Handle[T], fn func(context.Context, T) (U, error)) *MapHandle[T, U] {
	return &MapHandle[T, U]{
		fn:     fn,
		handle: getter,
	}
}
