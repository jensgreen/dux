package cancellable

import (
	"context"
	"errors"
	"fmt"
)

var ErrClosed = errors.New("closed")

func Send[T any](ctx context.Context, ch chan<- T, val T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- val:
		return nil
	}
}

func Receive[T any](ctx context.Context, ch <-chan T) (T, error) {
	select {
	case <-ctx.Done():
		var empty T
		return empty, ctx.Err()
	case val, ok := <-ch:
		if !ok {
			var empty T
			return empty, fmt.Errorf("channel of type %T is done: %w", empty, ErrClosed)
		}
		return val, nil
	}
}
