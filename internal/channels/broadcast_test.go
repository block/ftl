package channels

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestBroadcaster(t *testing.T) {
	t.Run("broadcasts messages to all subscribers", func(t *testing.T) {
		b := &Broadcaster[string]{}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create three subscribers
		sub1 := b.Subscribe()
		sub2 := b.Subscribe()
		sub3 := b.Subscribe()

		// Start the broadcaster
		go b.Run(ctx)

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
		b := &Broadcaster[int]{}
		ctx, cancel := context.WithCancel(context.Background())

		// Create subscribers
		sub1 := b.Subscribe()
		sub2 := b.Subscribe()

		// Start the broadcaster
		go b.Run(ctx)

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

	t.Run("stops broadcasting when context is cancelled", func(t *testing.T) {
		b := &Broadcaster[string]{}
		ctx, cancel := context.WithCancel(context.Background())

		// Start the broadcaster
		go b.Run(ctx)

		sub := b.Subscribe()
		cancel()

		// This broadcast should return immediately due to cancelled context
		b.Broadcast(ctx, "test")

		// Verify no message was sent
		select {
		case msg, ok := <-sub:
			if ok {
				t.Errorf("unexpected message received: %v", msg)
			}
		case <-time.After(100 * time.Millisecond):
			// Expected timeout
		}
	})
}
