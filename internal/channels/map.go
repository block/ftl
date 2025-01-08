package channels

func FlatMap[T any, U any](in <-chan T, fn func(T) []U) <-chan U {
	out := make(chan U)

	go func() {
		for n := range in {
			for _, u := range fn(n) {
				out <- u
			}
		}
		close(out)
	}()

	return out
}

func Map[T any, U any](in <-chan T, fn func(T) U) <-chan U {
	return FlatMap(in, func(t T) []U { return []U{fn(t)} })
}
