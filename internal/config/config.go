package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Runtime    RuntimeConfig
	Kubernetes KubernetesConfig
	Aliyun     AliyunConfig
	Sync       SyncConfig
}

type RuntimeConfig struct {
	AdapterMode string
}

type KubernetesConfig struct {
	SecretNamespace string
	SecretName      string
}

type AliyunConfig struct {
	Region           string
	CredentialSource string
	AccessKeyID      string
	AccessKeySecret  string
	CASEndpoint      string
	CDNEndpoint      string
	CDNDomains       []string
}

type SyncConfig struct {
	MaxRetries      int
	RetryBaseMillis int
	StateFile       string
}

func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	cfg, err := parseSimpleYAML(string(raw))
	if err != nil {
		return Config{}, err
	}
	applyEnvOverrides(&cfg)
	return cfg, nil
}

func (c Config) Validate() error {
	mode := strings.TrimSpace(c.Runtime.AdapterMode)
	if mode == "" {
		mode = "memory"
	}
	if mode != "memory" && mode != "api" {
		return errors.New("runtime.adapterMode must be memory or api")
	}

	if strings.TrimSpace(c.Kubernetes.SecretNamespace) == "" {
		return errors.New("kubernetes.secretNamespace is required")
	}
	if strings.TrimSpace(c.Kubernetes.SecretName) == "" {
		return errors.New("kubernetes.secretName is required")
	}
	if strings.TrimSpace(c.Aliyun.Region) == "" {
		return errors.New("aliyun.region is required")
	}
	if strings.TrimSpace(c.Aliyun.CredentialSource) == "" {
		return errors.New("aliyun.credentialSource is required")
	}
	if len(c.Aliyun.CDNDomains) == 0 {
		return errors.New("aliyun.cdnDomains must not be empty")
	}
	if mode == "api" {
		if strings.TrimSpace(c.Aliyun.CASEndpoint) == "" {
			return errors.New("aliyun.casEndpoint is required in api mode")
		}
		if strings.TrimSpace(c.Aliyun.CDNEndpoint) == "" {
			return errors.New("aliyun.cdnEndpoint is required in api mode")
		}
		if c.Aliyun.CredentialSource == "env" {
			if strings.Contains(c.Aliyun.AccessKeyID, "${") || strings.Contains(c.Aliyun.AccessKeySecret, "${") {
				return errors.New("aliyun.accessKeyId/accessKeySecret contain unresolved placeholders")
			}
			if strings.TrimSpace(c.Aliyun.AccessKeyID) == "" || strings.TrimSpace(c.Aliyun.AccessKeySecret) == "" {
				return errors.New("aliyun.accessKeyId and aliyun.accessKeySecret are required when credentialSource=env")
			}
		}
	}
	if c.Sync.MaxRetries < 0 {
		return errors.New("sync.maxRetries must be >= 0")
	}
	if c.Sync.RetryBaseMillis <= 0 {
		return errors.New("sync.retryBaseMillis must be > 0")
	}
	if strings.TrimSpace(c.Sync.StateFile) == "" {
		return errors.New("sync.stateFile is required")
	}

	return nil
}

func applyEnvOverrides(cfg *Config) {
	setString := func(env string, target *string) {
		v := strings.TrimSpace(os.Getenv(env))
		if v != "" {
			*target = v
		}
	}

	setString("CDN_CERT_SYNC_K8S_SECRET_NAMESPACE", &cfg.Kubernetes.SecretNamespace)
	setString("CDN_CERT_SYNC_K8S_SECRET_NAME", &cfg.Kubernetes.SecretName)
	setString("CDN_CERT_SYNC_ADAPTER_MODE", &cfg.Runtime.AdapterMode)
	setString("CDN_CERT_SYNC_ALIYUN_REGION", &cfg.Aliyun.Region)
	setString("CDN_CERT_SYNC_ALIYUN_CREDENTIAL_SOURCE", &cfg.Aliyun.CredentialSource)
	setString("CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_ID", &cfg.Aliyun.AccessKeyID)
	setString("CDN_CERT_SYNC_ALIYUN_ACCESS_KEY_SECRET", &cfg.Aliyun.AccessKeySecret)
	setString("CDN_CERT_SYNC_ALIYUN_CAS_ENDPOINT", &cfg.Aliyun.CASEndpoint)
	setString("CDN_CERT_SYNC_ALIYUN_CDN_ENDPOINT", &cfg.Aliyun.CDNEndpoint)

	if rawDomains := strings.TrimSpace(os.Getenv("CDN_CERT_SYNC_CDN_DOMAINS")); rawDomains != "" {
		parts := strings.Split(rawDomains, ",")
		domains := make([]string, 0, len(parts))
		for _, part := range parts {
			item := strings.TrimSpace(part)
			if item != "" {
				domains = append(domains, item)
			}
		}
		if len(domains) > 0 {
			cfg.Aliyun.CDNDomains = domains
		}
	}

	if rawRetries := strings.TrimSpace(os.Getenv("CDN_CERT_SYNC_MAX_RETRIES")); rawRetries != "" {
		if n, err := strconv.Atoi(rawRetries); err == nil {
			cfg.Sync.MaxRetries = n
		}
	}
	if rawBase := strings.TrimSpace(os.Getenv("CDN_CERT_SYNC_RETRY_BASE_MILLIS")); rawBase != "" {
		if n, err := strconv.Atoi(rawBase); err == nil {
			cfg.Sync.RetryBaseMillis = n
		}
	}
	setString("CDN_CERT_SYNC_STATE_FILE", &cfg.Sync.StateFile)
}
