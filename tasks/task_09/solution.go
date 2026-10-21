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

type Job[R any] struct {
	Value R
	Index int
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func Producer[T any](ctx context.Context, in []T) <-chan Job[T] {
	out := make(chan Job[T])

	go func() {
		defer close(out)
		for i, element := range in {
			select {
			case <-ctx.Done():
				return
			case out <- Job[T]{Value: element, Index: i}:
			}
		}
	}()

	return out
}

func Worker[T any, R any](
	ctx context.Context,
	jobs <-chan Job[T],
	fn func(context.Context, T) (R, error),
	results []Result[R],
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			value, err := fn(ctx, job.Value)

			res := Result[R]{
				Err: err,
			}

			if err == nil {
				res.Value = value
			}

			results[job.Index] = res
		}
	}
}

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]Result[R], error) {

	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(in) == 0 {
		return make([]Result[R], 0), nil
	}

	results := make([]Result[R], len(in))
	var wg sync.WaitGroup
	count := min(workers, len(in))
	jobs := Producer(ctx, in)

	for w := 0; w < count; w++ {
		wg.Go(func() {
			Worker(ctx, jobs, fn, results)
		})
	}

	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
