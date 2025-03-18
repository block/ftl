package channels

import (
	"context"
	"testing"
	"time"
)

func TestBroadcaster(t *testing.T) {
	t.Run("broadcasts messages to all subscribers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		b := NewNotifier(ctx)

		// Create three subscribers
		sub1 := b.Subscribe(ctx)
		sub2 := b.Subscribe(ctx)
		sub3 := b.Subscribe(ctx)

		// Send a message
		b.Notify(ctx)

		// Check that all subscribers received the message
		for _, ch := range []<-chan struct{}{sub1, sub2, sub3} {
			select {
			case <-ch:
			case <-ctx.Done():
				t.Error("timeout waiting for message")
			}
		}
	})

	t.Run("closes all subscriber channels when context is done", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b := NewNotifier(ctx)

		// Create subscribers
		sub1 := b.Subscribe(ctx)
		sub2 := b.Subscribe(ctx)

		// Cancel the context
		cancel()

		// Check that channels are closed
		for _, ch := range []<-chan struct{}{sub1, sub2} {
			select {
			case _, ok := <-ch:
				if ok {
					t.Error("channel should be closed")
				}
			case <-time.After(time.Second):
				t.Error("timeout waiting for channel to close")
			}
		}
	})

	t.Run("does not notify closed subscribers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b := NewNotifier(ctx)

		subCtx, subCancel := context.WithCancel(ctx)
		sub := b.Subscribe(subCtx)

		subCancel()

		b.Notify(ctx)

		select {
		case _, ok := <-sub:
			if ok {
				t.Errorf("subscriber should not receive message")
			}
		default:
		}
	})
}
