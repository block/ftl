package channels

import (
	"context"
	"slices"
	"sync"
)

type notifierSubscriber struct {
	ch  chan struct{}
	ctx context.Context
}

// Notifier helps to notify multiple subscribers that an event has occurred.
type Notifier struct {
	subscribers []*notifierSubscriber

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
func (b *Notifier) Subscribe(ctx context.Context) <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make(chan struct{}, 1)
	b.subscribers = append(b.subscribers, &notifierSubscriber{
		ch:  out,
		ctx: ctx,
	})

	return out
}

// Notify all subscribers.
func (b *Notifier) Notify(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()

	closedIndices := map[*notifierSubscriber]bool{}
	for _, sub := range b.subscribers {
		if sub.ctx.Err() != nil {
			// subscriber is closed, remove the subscriber
			closedIndices[sub] = true
			continue
		}
		if ctx.Err() != nil {
			// context is done, terminate
			return
		} else {
			select {
			case sub.ch <- struct{}{}:
			default:
				// channel is full, skip
			}
		}
	}

	for sub := range closedIndices {
		close(sub.ch)
	}
	b.subscribers = slices.DeleteFunc(b.subscribers, func(sub *notifierSubscriber) bool {
		return closedIndices[sub]
	})
}

func (b *Notifier) closeWhenDone(ctx context.Context) {
	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		for _, sub := range b.subscribers {
			close(sub.ch)
		}
	}()
}
