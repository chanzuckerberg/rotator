package sink

import (
	"context"
	"math/rand"
	"time"
)

const (
	defaultRetryAttempts = 5
	defaultRetrySleep    = time.Second
)

func retry(ctx context.Context, attempts int, sleep time.Duration, f func(context.Context) error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = f(ctx)
		if err == nil {
			return nil
		}

		jitter := time.Duration(rand.Int63n(int64(sleep)))
		time.Sleep(sleep + jitter)
	}
	return err
}
