package channels

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestBroadcaster(t *testing.T) {
	t.Run("broadcasts messages to all subscribers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		b := NewBroadcaster[string](ctx)

		// Create three subscribers
		sub1 := b.Subscribe()
		sub2 := b.Subscribe()
		sub3 := b.Subscribe()

		// Send a message
		msg := "hello"
		b.Broadcast(ctx, msg)

		// Check that all subscribers received the message
		for _, ch := range []<-chan string{sub1, sub2, sub3} {
			select {
			case received := <-ch:
				assert.Equal(t, msg, received)
			case <-ctx.Done():
				t.Error("timeout waiting for message")
			}
		}
	})

	t.Run("closes all subscriber channels when context is done", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b := NewBroadcaster[int](ctx)

		// Create subscribers
		sub1 := b.Subscribe()
		sub2 := b.Subscribe()

		// Cancel the context
		cancel()

		// Check that channels are closed
		for _, ch := range []<-chan int{sub1, sub2} {
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
}
