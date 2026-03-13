package aliyun

import "errors"

var (
	ErrRetryable = errors.New("aliyun retryable error")
	ErrTerminal  = errors.New("aliyun terminal error")
)
