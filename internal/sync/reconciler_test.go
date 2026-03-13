package sync

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/aliyun"
	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/config"
	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/k8s"
)

func TestReconcilerUploadsAndBinds(t *testing.T) {
	cfg := config.Config{
		Kubernetes: config.KubernetesConfig{
			SecretNamespace: "default",
			SecretName:      "site-cert",
		},
		Aliyun: config.AliyunConfig{
			Region:           "cn-hangzhou",
			CredentialSource: "env",
			CDNDomains:       []string{"a.example.com", "b.example.com"},
		},
		Sync: config.SyncConfig{
			MaxRetries:      1,
			RetryBaseMillis: 1,
			StateFile:       filepath.Join(t.TempDir(), "state-1.json"),
		},
	}

	stateStore, err := NewFileStateStore(cfg.Sync.StateFile)
	if err != nil {
		t.Fatalf("NewFileStateStore returned error: %v", err)
	}

	reconciler := NewReconciler(
		cfg,
		k8s.NewMemorySecretSource("default", "site-cert"),
		aliyun.NewMemoryCertificateStore(),
		aliyun.NewMemoryCDNBinder(),
		stateStore,
	)

	report, err := reconciler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if !report.Uploaded {
		t.Fatalf("expected uploaded=true")
	}
	if report.CertificateID == "" {
		t.Fatalf("expected non-empty certificate id")
	}
	if report.DomainsUpdated != 2 {
		t.Fatalf("expected domains updated = 2, got %d", report.DomainsUpdated)
	}
	if report.DomainFailures != 0 {
		t.Fatalf("expected domain failures = 0, got %d", report.DomainFailures)
	}
}

type retryableFlakySecretSource struct {
	failuresLeft int
}

func (f *retryableFlakySecretSource) GetTLSSecret(_ context.Context, namespace, name string) (k8s.TLSSecret, error) {
	if f.failuresLeft > 0 {
		f.failuresLeft--
		return k8s.TLSSecret{}, fmt.Errorf("temporary api issue: %w", k8s.ErrRetryable)
	}
	return k8s.TLSSecret{
		Namespace: namespace,
		Name:      name,
		CertPEM:   "-----BEGIN CERTIFICATE-----\nEXAMPLE\n-----END CERTIFICATE-----",
		KeyPEM:    "-----BEGIN PRIVATE KEY-----\nEXAMPLE\n-----END PRIVATE KEY-----",
	}, nil
}

func TestReconcilerRetriesOnRetryableSecretRead(t *testing.T) {
	cfg := config.Config{
		Kubernetes: config.KubernetesConfig{
			SecretNamespace: "default",
			SecretName:      "site-cert",
		},
		Aliyun: config.AliyunConfig{
			Region:           "cn-hangzhou",
			CredentialSource: "env",
			CDNDomains:       []string{"a.example.com"},
		},
		Sync: config.SyncConfig{
			MaxRetries:      2,
			RetryBaseMillis: 1,
			StateFile:       filepath.Join(t.TempDir(), "state-2.json"),
		},
	}

	stateStore, err := NewFileStateStore(cfg.Sync.StateFile)
	if err != nil {
		t.Fatalf("NewFileStateStore returned error: %v", err)
	}

	reconciler := NewReconciler(
		cfg,
		&retryableFlakySecretSource{failuresLeft: 1},
		aliyun.NewMemoryCertificateStore(),
		aliyun.NewMemoryCDNBinder(),
		stateStore,
	)

	report, err := reconciler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if report.Retries == 0 {
		t.Fatalf("expected retries > 0")
	}
}

func TestReconcilerSecondRunIsIdempotent(t *testing.T) {
	cfg := config.Config{
		Kubernetes: config.KubernetesConfig{
			SecretNamespace: "default",
			SecretName:      "site-cert",
		},
		Aliyun: config.AliyunConfig{
			Region:           "cn-hangzhou",
			CredentialSource: "env",
			CDNDomains:       []string{"a.example.com"},
		},
		Sync: config.SyncConfig{
			MaxRetries:      1,
			RetryBaseMillis: 1,
			StateFile:       filepath.Join(t.TempDir(), "state-3.json"),
		},
	}

	stateStore, err := NewFileStateStore(cfg.Sync.StateFile)
	if err != nil {
		t.Fatalf("NewFileStateStore returned error: %v", err)
	}

	store := aliyun.NewMemoryCertificateStore()
	reconciler := NewReconciler(
		cfg,
		k8s.NewMemorySecretSource("default", "site-cert"),
		store,
		aliyun.NewMemoryCDNBinder(),
		stateStore,
	)

	first, err := reconciler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("first RunOnce returned error: %v", err)
	}
	second, err := reconciler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("second RunOnce returned error: %v", err)
	}

	if !first.Uploaded {
		t.Fatalf("expected first run uploaded=true")
	}
	if second.Uploaded {
		t.Fatalf("expected second run uploaded=false")
	}
	if first.CertificateID != second.CertificateID {
		t.Fatalf("expected stable certificate id, first=%q second=%q", first.CertificateID, second.CertificateID)
	}
}
