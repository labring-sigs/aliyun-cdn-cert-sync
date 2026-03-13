package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLAndEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `
runtime:
  adapterMode: api
kubernetes:
  secretNamespace: cert-manager
  secretName: site-tls
aliyun:
  region: cn-hangzhou
  credentialSource: env
  accessKeyId: test-ak
  accessKeySecret: test-sk
  casEndpoint: https://cas.aliyuncs.com
  cdnEndpoint: https://cdn.aliyuncs.com
  cdnDomains:
    - a.example.com
sync:
  maxRetries: 1
  retryBaseMillis: 100
  stateFile: /tmp/certmap.json
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	t.Setenv("CDN_CERT_SYNC_CDN_DOMAINS", "x.example.com, y.example.com")
	t.Setenv("CDN_CERT_SYNC_MAX_RETRIES", "5")
	t.Setenv("CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_ID", "env-ak")
	t.Setenv("CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_SECRET", "env-sk")
	t.Setenv("CDN_CERT_SYNC_ALIYUN_CAS_ENDPOINT", "https://cas.aliyuncs.com")
	t.Setenv("CDN_CERT_SYNC_ALIYUN_CDN_ENDPOINT", "https://cdn.aliyuncs.com")
	t.Setenv("CDN_CERT_SYNC_STATE_FILE", "/tmp/state-from-env.json")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Sync.MaxRetries != 5 {
		t.Fatalf("expected maxRetries=5, got %d", cfg.Sync.MaxRetries)
	}
	if len(cfg.Aliyun.CDNDomains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(cfg.Aliyun.CDNDomains))
	}
	if cfg.Aliyun.CDNDomains[0] != "x.example.com" {
		t.Fatalf("unexpected first domain: %s", cfg.Aliyun.CDNDomains[0])
	}
	if cfg.Sync.StateFile != "/tmp/state-from-env.json" {
		t.Fatalf("unexpected state file: %s", cfg.Sync.StateFile)
	}
	if cfg.Runtime.AdapterMode != "api" {
		t.Fatalf("expected adapter mode api, got %s", cfg.Runtime.AdapterMode)
	}
}
