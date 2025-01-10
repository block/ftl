package iterops

import "iter"

// Empty returns an empty iterator.
func Empty[T any]() iter.Seq[T] {
	return func(yield func(T) bool) {}
}
