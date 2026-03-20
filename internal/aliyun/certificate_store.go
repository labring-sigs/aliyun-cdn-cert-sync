package aliyun

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Certificate struct {
	ID          string
	Fingerprint string
	CertPEM     string
	KeyPEM      string
}

type CertificateStore interface {
	FindByFingerprint(ctx context.Context, fingerprint, resourceGroupID string) (Certificate, error)
	Create(ctx context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error)
}

type CASAPI interface {
	FindCertificateByFingerprint(ctx context.Context, fingerprint, resourceGroupID string) (Certificate, error)
	UploadCertificate(ctx context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error)
}

type CASCertificateStore struct {
	api CASAPI
}

func NewCASCertificateStore(api CASAPI) *CASCertificateStore {
	return &CASCertificateStore{api: api}
}

func (s *CASCertificateStore) FindByFingerprint(ctx context.Context, fingerprint, resourceGroupID string) (Certificate, error) {
	return s.api.FindCertificateByFingerprint(ctx, fingerprint, resourceGroupID)
}

func (s *CASCertificateStore) Create(ctx context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error) {
	return s.api.UploadCertificate(ctx, certPEM, keyPEM, fingerprint)
}

type MemoryCertificateStore struct {
	mu      sync.Mutex
	records map[string]Certificate
	nextID  int
}

func NewMemoryCertificateStore() *MemoryCertificateStore {
	return &MemoryCertificateStore{
		records: make(map[string]Certificate),
		nextID:  1,
	}
}

func (s *MemoryCertificateStore) FindByFingerprint(_ context.Context, fingerprint, resourceGroupID string) (Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[fingerprint]
	if !ok {
		return Certificate{}, ErrNotFound
	}

	return record, nil
}

func (s *MemoryCertificateStore) Create(_ context.Context, certPEM, keyPEM, fingerprint string) (Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.records[fingerprint]; ok {
		return existing, nil
	}

	record := Certificate{
		ID:          fmt.Sprintf("cas-cert-%d", s.nextID),
		Fingerprint: fingerprint,
		CertPEM:     certPEM,
		KeyPEM:      keyPEM,
	}
	s.records[fingerprint] = record
	s.nextID++

	return record, nil
}
