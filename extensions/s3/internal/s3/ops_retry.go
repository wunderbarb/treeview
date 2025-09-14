package s3

import (
	"context"
	"errors"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
)

var errFailedFailSafe = errors.New("failed ")

func hasAccessWithRetry(ctx context.Context, path string, opts ...Option) bool {
	return failsafe.Run(func() error {
		if hasAccess(ctx, path, opts...) {
			return nil
		}
		return errFailedFailSafe
	}, getRetryPolicy(opts...)) == nil
}

func getRetryPolicy(opts ...Option) retrypolicy.RetryPolicy[any] {
	o := collectOptions(opts...)
	retryPolicy := retrypolicy.NewBuilder[any]().WithMaxRetries(o.retryMaxAttempt).WithDelay(o.retryLapse).Build()
	return retryPolicy
}
