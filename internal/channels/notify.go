package channels

import (
	"context"
	"sync"
)

// Notifier helps to notify multiple subscribers that an event has occurred.
type Notifier struct {
	subscribers []chan struct{}

	mu sync.Mutex
}

// NewNotifier creates a new Notifier instance.
//
// The returned Notifier will automatically close all subscribers when the context is cancelled.
func NewNotifier(ctx context.Context) *Notifier {
	b := &Notifier{}
	b.closeWhenDone(ctx)
	return b
}

// Subscribe to the notifier.
//
// The returned channel will be closed when the notifier context is cancelled.
func (b *Notifier) Subscribe() <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make(chan struct{}, 1)
	b.subscribers = append(b.subscribers, out)

	return out
}

// Notify all subscribers.
func (b *Notifier) Notify(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, ch := range b.subscribers {
		select {
		case <-ctx.Done():
			return
		case ch <- struct{}{}:
		default:
			// channel is full, skip
		}
	}
}

func (b *Notifier) closeWhenDone(ctx context.Context) {
	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for _, ch := range b.subscribers {
			close(ch)
		}
	}()
}
