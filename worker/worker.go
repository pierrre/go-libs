package worker

import (
	"context"
	"time"
)

type Func func(ctx context.Context)

func Run(ctx context.Context, f Func, period time.Duration) {
	var ticker *time.Ticker
	if period > 0 {
		ticker = time.NewTicker(period)
		defer ticker.Stop()

	}
	for ctx.Err() == nil {
		f(ctx)
		if ticker != nil {
			select {
			case <-ticker.C:
			case <-ctx.Done():
			}
		}
	}
}

type FuncError func(ctx context.Context) error

func WithError(ctx context.Context, f FuncError, onError func(ctx context.Context, err error)) Func {
	return func(ctx context.Context) {
		err := f(ctx)
		if err != nil {
			onError(ctx, err)
		}
	}
}
