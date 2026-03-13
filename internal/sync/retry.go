package sync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/aliyun"
	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/k8s"
)

func withRetry(
	ctx context.Context,
	maxRetries int,
	baseDelay time.Duration,
	fn func() error,
) (int, error) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}

	var retries int
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return retries, nil
		}
		if !isRetryable(err) {
			return retries, err
		}
		if attempt == maxRetries {
			return retries, fmt.Errorf("retry limit exceeded: %w", err)
		}

		retries++
		sleep := baseDelay * time.Duration(1<<attempt)
		select {
		case <-ctx.Done():
			return retries, ctx.Err()
		case <-time.After(sleep):
		}
	}

	return retries, nil
}

func isRetryable(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, aliyun.ErrRetryable) ||
		errors.Is(err, k8s.ErrRetryable)
}
