package iterops

import "iter"

func Concat[T any](in ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, n := range in {
			for m := range n {
				if !yield(m) {
					return
				}
			}
		}
	}
}
