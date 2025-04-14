package iterops

import "iter"

func Filter[T any](in iter.Seq[T], pred func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		in(func(n T) bool {
			if pred(n) {
				return yield(n)
			}
			return true
		})
	}
}
