package k8s

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

type TLSSecret struct {
	Namespace string
	Name      string
	CertPEM   string
	KeyPEM    string
}

func (s TLSSecret) Fingerprint() (string, error) {
	if strings.TrimSpace(s.CertPEM) == "" {
		return "", errors.New("empty cert pem")
	}

	block, _ := pem.Decode([]byte(s.CertPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return "", errors.New("invalid cert pem")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse cert pem: %w", err)
	}

	sum := sha1.Sum(cert.Raw)
	hexDigest := strings.ToUpper(hex.EncodeToString(sum[:]))

	var builder strings.Builder
	builder.Grow(len(hexDigest) + (len(sum) - 1))
	for i := 0; i < len(hexDigest); i += 2 {
		if i > 0 {
			builder.WriteByte(':')
		}
		builder.WriteString(hexDigest[i : i+2])
	}

	return builder.String(), nil
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
