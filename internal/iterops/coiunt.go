package iterops

import "iter"

func Count[T any](in iter.Seq[T], pred func(T) bool) int {
	return Len(Filter(in, pred))
}
