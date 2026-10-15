package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	next := 0
	results := make([]Result[R], len(in))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < min(workers, len(in)); i++ {
		wg.Go(func() {
			for {
				mu.Lock()
				index := next

				if ctx.Err() != nil || index >= len(in) {
					mu.Unlock()
					return
				}

				next++
				mu.Unlock()

				resp, err := fn(ctx, in[index])

				results[index] = Result[R]{Err: err}
				if err == nil {
					results[index].Value = resp
				}

			}
		})
	}

	wg.Wait()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return results, nil
}
