package channels

import (
	"context"
	"sync"
)

// Broadcaster helps to broadcast messages to multiple subscribed channels.
type Broadcaster[T any] struct {
	subscribers []chan T

	mu sync.Mutex
}

// NewBroadcaster creates a new Broadcaster instance.
//
// The returned Broadcaster will automatically close all subscribers when the context is cancelled.
func NewBroadcaster[T any](ctx context.Context) *Broadcaster[T] {
	b := &Broadcaster[T]{}
	b.closeWhenDone(ctx)
	return b
}

// Subscribe to the broadcaster.
//
// The returned channel will be closed when the broadcaster context is cancelled.
func (b *Broadcaster[T]) Subscribe() <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make(chan T, 16)
	b.subscribers = append(b.subscribers, out)

	return out
}

// Broadcast a message to all subscribers.
func (b *Broadcaster[T]) Broadcast(ctx context.Context, msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.subscribers {
		select {
		case <-ctx.Done():
			return
		case ch <- msg:
		default:
			// channel is full, skip
		}
	}
}

func (b *Broadcaster[T]) closeWhenDone(ctx context.Context) {
	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for _, ch := range b.subscribers {
			close(ch)
		}
	}()
}
