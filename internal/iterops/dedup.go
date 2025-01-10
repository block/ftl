package iterops

import (
	"iter"
	"reflect"
)

// Dedup returns an iterator that yields values from the input iterator, removing consecutive duplicates.
func Dedup[T any](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var last T
		seq(func(v T) bool {
			if reflect.DeepEqual(v, last) {
				return true
			}
			last = v
			return yield(v)
		})
	}
}
