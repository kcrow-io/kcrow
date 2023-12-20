package util

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Retfn func() error

func Backoff(rfn Retfn) error {
	newbo := backoff.WithMaxRetries(&backoff.ConstantBackOff{Interval: time.Microsecond * 10}, 3)
	return backoff.Retry(backoff.Operation(rfn), newbo)
}

func TimeBackoff(rfn Retfn) error {
	expbf := backoff.NewExponentialBackOff()
	expbf.InitialInterval = time.Second * 1
	expbf.MaxElapsedTime = time.Second * 30

	return backoff.Retry(backoff.Operation(rfn), expbf)
}
