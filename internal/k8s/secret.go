package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type TLSSecret struct {
	Namespace string
	Name      string
	CertPEM   string
	KeyPEM    string
}

func (s TLSSecret) Fingerprint() (string, error) {
	if s.CertPEM == "" {
		return "", errors.New("empty cert pem")
	}
	sum := sha256.Sum256([]byte(s.CertPEM))
	return hex.EncodeToString(sum[:]), nil
}

type SecretSource interface {
	GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error)
}

type API interface {
	GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error)
}

type APISecretSource struct {
	api API
}

func NewAPISecretSource(api API) *APISecretSource {
	return &APISecretSource{api: api}
}

func (s *APISecretSource) GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error) {
	return s.api.GetTLSSecret(ctx, namespace, name)
}

type ClientConfig struct {
	InCluster  bool
	Kubeconfig string
}

type Client struct {
	impl API
}

func (c *Client) GetTLSSecret(ctx context.Context, namespace, name string) (TLSSecret, error) {
	return c.impl.GetTLSSecret(ctx, namespace, name)
}

type MemorySecretSource struct {
	secret TLSSecret
}

func NewMemorySecretSource(namespace, name string) *MemorySecretSource {
	return &MemorySecretSource{
		secret: TLSSecret{
			Namespace: namespace,
			Name:      name,
			CertPEM:   "-----BEGIN CERTIFICATE-----\nEXAMPLE\n-----END CERTIFICATE-----",
			KeyPEM:    "-----BEGIN PRIVATE KEY-----\nEXAMPLE\n-----END PRIVATE KEY-----",
		},
	}
}

func (s *MemorySecretSource) GetTLSSecret(_ context.Context, namespace, name string) (TLSSecret, error) {
	if namespace != s.secret.Namespace || name != s.secret.Name {
		return TLSSecret{}, fmt.Errorf("secret not found: %s/%s", namespace, name)
	}
	if s.secret.CertPEM == "" || s.secret.KeyPEM == "" {
		return TLSSecret{}, errors.New("secret missing tls.crt or tls.key")
	}

	return s.secret, nil
}
