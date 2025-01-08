package channels

import (
	"context"
	"sync"
)

// Broadcaster creates a channel that broadcasts messages to all subscribers.
type Broadcaster[T any] struct {
	subscribers []chan T

	mu sync.Mutex
}

func (b *Broadcaster[T]) Subscribe() <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make(chan T, 16)
	b.subscribers = append(b.subscribers, out)

	return out
}

func (b *Broadcaster[T]) Run(ctx context.Context) {
	<-ctx.Done()

	for _, ch := range b.subscribers {
		close(ch)
	}
}

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
