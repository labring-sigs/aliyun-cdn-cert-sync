package k8s

import "errors"

var (
	ErrRetryable = errors.New("k8s retryable error")
	ErrTerminal  = errors.New("k8s terminal error")
)
