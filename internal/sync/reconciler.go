package sync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/aliyun"
	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/config"
	"github.com/allosaurus/aliyun-cdn-cert-sync/internal/k8s"
)

type Report struct {
	Uploaded       bool
	CertificateID  string
	DomainsUpdated int
	DomainFailures int
	Retries        int
}

type Reconciler struct {
	cfg        config.Config
	secrets    k8s.SecretSource
	certs      aliyun.CertificateStore
	cdnBinder  aliyun.CDNBinder
	stateStore StateStore
}

func NewReconciler(
	cfg config.Config,
	secrets k8s.SecretSource,
	certs aliyun.CertificateStore,
	cdnBinder aliyun.CDNBinder,
	stateStore StateStore,
) *Reconciler {
	return &Reconciler{
		cfg:        cfg,
		secrets:    secrets,
		certs:      certs,
		cdnBinder:  cdnBinder,
		stateStore: stateStore,
	}
}

func (r *Reconciler) RunOnce(ctx context.Context) (Report, error) {
	var report Report
	baseDelay := time.Duration(r.cfg.Sync.RetryBaseMillis) * time.Millisecond

	var secret k8s.TLSSecret
	readRetries, err := withRetry(ctx, r.cfg.Sync.MaxRetries, baseDelay, func() error {
		var getErr error
		secret, getErr = r.secrets.GetTLSSecret(ctx, r.cfg.Kubernetes.SecretNamespace, r.cfg.Kubernetes.SecretName)
		return getErr
	})
	report.Retries += readRetries
	if err != nil {
		return report, fmt.Errorf("read kubernetes tls secret: %w", err)
	}

	fingerprint, err := secret.Fingerprint()
	if err != nil {
		return report, fmt.Errorf("build cert fingerprint: %w", err)
	}

	certificate, uploaded, retries, err := r.ensureCertificate(ctx, secret, fingerprint, r.cfg.Aliyun.ResourceGroupID, baseDelay)
	report.Uploaded = uploaded
	report.Retries += retries
	if err != nil {
		return report, err
	}
	report.CertificateID = certificate.ID

	for _, domain := range r.cfg.Aliyun.CDNDomains {
		bindRetries, bindErr := withRetry(ctx, r.cfg.Sync.MaxRetries, baseDelay, func() error {
			return r.cdnBinder.BindCertificate(ctx, domain, certificate.ID)
		})
		report.Retries += bindRetries
		if bindErr != nil {
			report.DomainFailures++
			continue
		}
		report.DomainsUpdated++
	}

	return report, nil
}

func (r *Reconciler) ensureCertificate(
	ctx context.Context,
	secret k8s.TLSSecret,
	fingerprint string,
	resourceGroupID string,
	baseDelay time.Duration,
) (aliyun.Certificate, bool, int, error) {
	var (
		certificate aliyun.Certificate
		uploaded    bool
	)
	retries, err := withRetry(ctx, r.cfg.Sync.MaxRetries, baseDelay, func() error {
		var runErr error
		certificate, uploaded, runErr = r.ensureCertificateOnce(ctx, secret, fingerprint, resourceGroupID)
		return runErr
	})
	if err != nil {
		return aliyun.Certificate{}, false, retries, fmt.Errorf("ensure certificate: %w", err)
	}
	return certificate, uploaded, retries, nil
}

func (r *Reconciler) ensureCertificateOnce(
	ctx context.Context,
	secret k8s.TLSSecret,
	fingerprint, resourceGroupID string,
) (aliyun.Certificate, bool, error) {
	if r.stateStore != nil {
		if certID, ok, err := r.stateStore.GetCertIDByFingerprint(fingerprint); err == nil && ok && certID != "" {
			return aliyun.Certificate{
				ID:          certID,
				Fingerprint: fingerprint,
			}, false, nil
		}
	}

	existing, err := r.certs.FindByFingerprint(ctx, fingerprint, resourceGroupID)
	if err == nil {
		if r.stateStore != nil {
			_ = r.stateStore.SetCertIDByFingerprint(fingerprint, existing.ID)
		}
		return existing, false, nil
	}
	if !errors.Is(err, aliyun.ErrNotFound) {
		return aliyun.Certificate{}, false, fmt.Errorf("query certificate store: %w", err)
	}

	created, err := r.certs.Create(ctx, secret.CertPEM, secret.KeyPEM, fingerprint)
	if err != nil {
		return aliyun.Certificate{}, false, fmt.Errorf("create certificate in aliyun: %w", err)
	}
	if r.stateStore != nil {
		_ = r.stateStore.SetCertIDByFingerprint(fingerprint, created.ID)
	}
	return created, true, nil
}
