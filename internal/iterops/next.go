package iterops

import (
	"iter"
)

func Next[T any](in iter.Seq[T]) (T, bool) {
	for v := range in {
		return v, true
	}
	return *new(T), false
}
