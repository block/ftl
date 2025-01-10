package iterops

import (
	"iter"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/tuple"
)

// WindowPair returns a window of size 2 of the input iterator.
func WindowPair[T any](in iter.Seq[T]) iter.Seq[tuple.Pair[T, T]] {
	return func(yield func(tuple.Pair[T, T]) bool) {
		previous := optional.None[T]()
		for n := range in {
			if val, ok := previous.Get(); ok {
				result := tuple.PairOf(val, n)
				if !yield(result) {
					return
				}
			}
			previous = optional.Some(n)
		}
	}
}
