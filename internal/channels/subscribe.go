package channels

import (
	"context"
	"iter"
)

// Subscribable represents an object that can be subscribed too until the context is cancelled
type Subscribable[T any] interface {
	Subscribe(ctx context.Context) <-chan T
}

// IterSubscribable subscribes to the object and iterates over events until the context is cancelled
//
// Check ctx.Err() != nil to detect if the context was cancelled.
func IterSubscribable[T any](ctx context.Context, subscribable Subscribable[T]) iter.Seq[T] {
	return IterContext(ctx, subscribable.Subscribe(ctx))
}
