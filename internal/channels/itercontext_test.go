package channels

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestIterContext(t *testing.T) {
	t.Run("iterates until channel closed", func(t *testing.T) {
		ch := make(chan int)
		ctx := context.Background()

		// Start goroutine to send values
		go func() {
			ch <- 1
			ch <- 2
			ch <- 3
			close(ch)
		}()

		assert.Equal(t, []int{1, 2, 3}, slices.Collect(IterContext(ctx, ch)))
	})

	t.Run("stops when context cancelled", func(t *testing.T) {
		ch := make(chan int)
		ctx, cancel := context.WithCancel(context.Background())

		// Start goroutine to send values
		go func() {
			ch <- 1
			ch <- 2
			time.Sleep(10 * time.Millisecond) // Small delay to ensure cancel happens
			cancel()                          // Cancel context before sending 3
			ch <- 3                           // This should not be received
			close(ch)
		}()

		assert.Equal(t, []int{1, 2}, slices.Collect(IterContext(ctx, ch)))
		assert.Error(t, ctx.Err())
	})
}
