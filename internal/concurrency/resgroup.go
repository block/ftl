package concurrency

type ResourceGroup[T any] struct {
	funcs []func() (T, error)

	reschan chan T
	errchan chan error
}

func (r *ResourceGroup[T]) Go(f func() (T, error)) {
	if r.reschan == nil {
		r.reschan = make(chan T, 64)
		r.errchan = make(chan error, 64)
	}

	r.funcs = append(r.funcs, f)

	go func() {
		res, err := f()
		r.reschan <- res
		r.errchan <- err
	}()
}

func (r *ResourceGroup[T]) Wait() ([]T, error) {
	results := make([]T, 0, len(r.funcs))
	for range r.funcs {
		err := <-r.errchan
		if err != nil {
			return nil, err
		}
		results = append(results, <-r.reschan)
	}
	return results, nil
}
