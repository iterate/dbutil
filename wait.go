package pgutil

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

func Wait(ctx context.Context, pinger Pinger) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 250 * time.Millisecond
	b := backoff.WithContext(bo, ctx)

	err := backoff.Retry(func() error {
		return pinger.PingContext(ctx)
	}, b)
	if err != nil {
		// Handle error.
		return err
	}

	return nil
}
