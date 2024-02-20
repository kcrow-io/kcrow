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

// maxtime is 0 will forever
func TimeBackoff(rfn Retfn, maxtime time.Duration) error {
	expbf := backoff.NewExponentialBackOff()
	expbf.InitialInterval = time.Second * 1
	expbf.MaxElapsedTime = maxtime

	return backoff.Retry(backoff.Operation(rfn), expbf)
}
