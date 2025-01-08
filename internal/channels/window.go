package channels

import (
	"github.com/alecthomas/types/tuple"
)

// WindowPair returns window of size 2 of the input channel.
func WindowPair[T any](in <-chan T) <-chan tuple.Pair[T, T] {
	out := make(chan tuple.Pair[T, T])

	var previous T
	go func() {
		for n := range in {
			pair := tuple.PairOf(previous, n)
			previous = n
			out <- pair
		}
		close(out)
	}()

	return out
}
