package iterops

import "iter"

func Take[T any](in iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for x := range in {
			if count >= n {
				return
			}
			count++
			if !yield(x) {
				return
			}
		}
	}
}
