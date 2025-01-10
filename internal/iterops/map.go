package iterops

import "iter"

func Map[T any, U any](in iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for n := range in {
			if !yield(fn(n)) {
				return
			}
		}
	}
}

func FlatMap[T any, U any](in iter.Seq[T], fn func(T) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		for n := range in {
			for u := range fn(n) {
				if !yield(u) {
					return
				}
			}
		}
	}
}
