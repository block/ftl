package iterops

import "iter"

func Const[T any](in ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, n := range in {
			if !yield(n) {
				return
			}
		}
	}
}
