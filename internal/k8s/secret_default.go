//go:build !clientgo

package k8s

import (
	"context"
	"fmt"
	"strings"
)

type stubClient struct{}

func NewClient(cfg ClientConfig) (*Client, error) {
	if !cfg.InCluster && strings.TrimSpace(cfg.Kubeconfig) == "" {
		return nil, fmt.Errorf("%w: kubeconfig is required when not in-cluster", ErrTerminal)
	}
	return &Client{impl: &stubClient{}}, nil
}

func (c *stubClient) GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error) {
	if strings.TrimSpace(namespace) == "" || strings.TrimSpace(name) == "" {
		return TLSSecret{}, fmt.Errorf("%w: namespace/name is required", ErrTerminal)
	}
	if err := ctx.Err(); err != nil {
		return TLSSecret{}, fmt.Errorf("%w: context done", ErrRetryable)
	}
	return TLSSecret{}, fmt.Errorf("%w: build with -tags=clientgo to enable Kubernetes API integration", ErrTerminal)
}
