package channels

import (
	"time"
)

// Take a value from the given channel, calling Fatalf if it is not available within the timeout.
func Take[T any](t interface{ Fatalf(string, ...any) }, timeout time.Duration, ch chan T) (out T) { //nolint
	select {
	case out = <-ch:
		return out

	case <-time.After(timeout):
		t.Fatalf("timed out waiting for %T", out)
		return
	}
}
