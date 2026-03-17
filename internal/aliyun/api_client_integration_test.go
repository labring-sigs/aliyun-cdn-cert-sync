//go:build integration

package aliyun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAPIClientLiveCASOperations(t *testing.T) {
	cfg := loadLiveAPIClientConfig(t)
	if cfg == nil {
		t.Skip("live Aliyun CAS test config not found")
	}

	client, err := NewAPIClient(cfg.APIClientConfig)
	if err != nil {
		t.Fatalf("NewAPIClient returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cert, err := client.FindCertificateByFingerprint(ctx, cfg.LiveCertificateFingerprint)
	if err != nil {
		t.Fatalf("FindCertificateByFingerprint returned error: %v", err)
	}
	if cert.ID == "" {
		t.Fatal("expected non-empty certificate id")
	}
}

func TestAPIClientLiveCDNBinding(t *testing.T) {
	cfg := loadLiveAPIClientConfig(t)
	if cfg == nil {
		t.Skip("live Aliyun CDN test config not found")
	}
	if cfg.LiveCertificateID == "" || cfg.LiveCDNDomain == "" {
		t.Skip("live CDN binding test requires liveCertificateId and liveCdnDomain")
	}

	client, err := NewAPIClient(cfg.APIClientConfig)
	if err != nil {
		t.Fatalf("NewAPIClient returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.UpdateDomainCertificate(ctx, cfg.LiveCDNDomain, cfg.LiveCertificateID); err != nil {
		t.Fatalf("UpdateDomainCertificate returned error: %v", err)
	}
}

type liveAPIClientConfig struct {
	APIClientConfig
	LiveCertificateFingerprint string
	LiveCertificateID          string
	LiveCDNDomain              string
}

func loadLiveAPIClientConfig(t *testing.T) *liveAPIClientConfig {
	t.Helper()

	path := filepath.Join("testdata", "aliyun-live.env")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		t.Fatalf("read live test config: %v", err)
	}

	values := make(map[string]string)
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid live test config line: %q", line)
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	cfg := &liveAPIClientConfig{
		APIClientConfig: APIClientConfig{
			Region:           values["ALIYUN_REGION"],
			CredentialSource: "env",
			AccessKeyID:      values["ALIYUN_ACCESS_KEY_ID"],
			AccessKeySecret:  values["ALIYUN_ACCESS_KEY_SECRET"],
			CASEndpoint:      values["ALIYUN_CAS_ENDPOINT"],
			CDNEndpoint:      values["ALIYUN_CDN_ENDPOINT"],
		},
		LiveCertificateFingerprint: values["ALIYUN_LIVE_CERT_FINGERPRINT"],
		LiveCertificateID:          values["ALIYUN_LIVE_CERT_ID"],
		LiveCDNDomain:              values["ALIYUN_LIVE_CDN_DOMAIN"],
	}

	required := map[string]string{
		"ALIYUN_REGION":                cfg.Region,
		"ALIYUN_ACCESS_KEY_ID":         cfg.AccessKeyID,
		"ALIYUN_ACCESS_KEY_SECRET":     cfg.AccessKeySecret,
		"ALIYUN_CAS_ENDPOINT":          cfg.CASEndpoint,
		"ALIYUN_CDN_ENDPOINT":          cfg.CDNEndpoint,
		"ALIYUN_LIVE_CERT_FINGERPRINT": cfg.LiveCertificateFingerprint,
	}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("missing required live test config value %s in %s", key, path)
		}
	}

	return cfg
}
