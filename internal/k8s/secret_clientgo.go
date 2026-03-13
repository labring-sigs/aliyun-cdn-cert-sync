//go:build clientgo

package k8s

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type clientsetSecretClient struct {
	client kubernetes.Interface
}

func NewClient(cfg ClientConfig) (*Client, error) {
	var restConfig *rest.Config
	var err error
	if cfg.InCluster {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: build kube config: %v", ErrTerminal, err)
	}

	cs, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: init clientset: %v", ErrTerminal, err)
	}

	return &Client{
		impl: &clientsetSecretClient{
			client: cs,
		},
	}, nil
}

func (c *clientsetSecretClient) GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error) {
	secret, err := c.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) || apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
			return TLSSecret{}, fmt.Errorf("%w: %v", ErrTerminal, err)
		}
		return TLSSecret{}, fmt.Errorf("%w: %v", ErrRetryable, err)
	}

	cert := strings.TrimSpace(string(secret.Data["tls.crt"]))
	key := strings.TrimSpace(string(secret.Data["tls.key"]))
	if cert == "" || key == "" {
		return TLSSecret{}, fmt.Errorf("%w: secret missing tls.crt or tls.key", ErrTerminal)
	}

	return TLSSecret{
		Namespace: namespace,
		Name:      name,
		CertPEM:   cert,
		KeyPEM:    key,
	}, nil
}
