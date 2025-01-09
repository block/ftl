package iterops

import (
	"iter"

	"github.com/alecthomas/types/tuple"
)

// WindowPair returns a window of size 2 of the input iterator.
func WindowPair[T any](in iter.Seq[T], start T) iter.Seq[tuple.Pair[T, T]] {
	return func(yield func(tuple.Pair[T, T]) bool) {
		previous := start
		for n := range in {
			result := tuple.PairOf(previous, n)
			previous = n
			if !yield(result) {
				return
			}
		}
	}
}
