package iterops

import "iter"

func Len[T any](in iter.Seq[T]) int {
	var out int
	in(func(T) bool {
		out++
		return true
	})
	return out
}
