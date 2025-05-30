package channels

import (
	"context"
	"iter"
)

// Subscribeable represents an object that can be subscribed too until the context is cancelled
type Subscribeable[T any] interface {
	Subscribe(ctx context.Context) <-chan T
}

// IterSubscribeable subscribes to the object and iterates over events until the context is cancelled
//
// Check ctx.Err() != nil to detect if the context was cancelled.
func IterSubscribeable[T any](ctx context.Context, subscribeable Subscribeable[T]) iter.Seq[T] {
	return IterContext(ctx, subscribeable.Subscribe(ctx))
}
