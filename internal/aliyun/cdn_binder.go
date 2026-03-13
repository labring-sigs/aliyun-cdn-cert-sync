package aliyun

import (
	"context"
	"fmt"
	"sync"
)

type CDNBinder interface {
	BindCertificate(ctx context.Context, domain, certificateID string) error
}

type CDNAPI interface {
	UpdateDomainCertificate(ctx context.Context, domain, certificateID string) error
}

type APICDNBinder struct {
	api CDNAPI
}

func NewAPICDNBinder(api CDNAPI) *APICDNBinder {
	return &APICDNBinder{api: api}
}

func (b *APICDNBinder) BindCertificate(ctx context.Context, domain, certificateID string) error {
	return b.api.UpdateDomainCertificate(ctx, domain, certificateID)
}

type MemoryCDNBinder struct {
	mu      sync.Mutex
	binding map[string]string
}

func NewMemoryCDNBinder() *MemoryCDNBinder {
	return &MemoryCDNBinder{
		binding: make(map[string]string),
	}
}

func (b *MemoryCDNBinder) BindCertificate(_ context.Context, domain, certificateID string) error {
	if domain == "" {
		return fmt.Errorf("domain is empty")
	}
	if certificateID == "" {
		return fmt.Errorf("certificate id is empty")
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.binding[domain] = certificateID
	return nil
}
