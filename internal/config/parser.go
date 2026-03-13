package config

import (
	"fmt"
	"strconv"
	"strings"
)

func parseSimpleYAML(content string) (Config, error) {
	cfg := Config{
		Runtime: RuntimeConfig{
			AdapterMode: "memory",
		},
		Sync: SyncConfig{
			MaxRetries:      3,
			RetryBaseMillis: 200,
			StateFile:       "./state/certmap.json",
		},
	}

	lines := strings.Split(content, "\n")
	var section string
	inDomains := false
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "- ") {
			key := strings.TrimSuffix(line, ":")
			if key == "cdnDomains" && section == "aliyun" {
				inDomains = true
				continue
			}
			section = key
			inDomains = false
			continue
		}

		if strings.HasPrefix(line, "- ") {
			if section == "aliyun" && inDomains {
				cfg.Aliyun.CDNDomains = append(cfg.Aliyun.CDNDomains, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
				continue
			}
			return Config{}, fmt.Errorf("unexpected list item outside supported section: %q", line)
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return Config{}, fmt.Errorf("invalid line: %q", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		inDomains = false

		switch section {
		case "runtime":
			switch key {
			case "adapterMode":
				cfg.Runtime.AdapterMode = value
			default:
				return Config{}, fmt.Errorf("unknown runtime key: %s", key)
			}
		case "kubernetes":
			switch key {
			case "secretNamespace":
				cfg.Kubernetes.SecretNamespace = value
			case "secretName":
				cfg.Kubernetes.SecretName = value
			default:
				return Config{}, fmt.Errorf("unknown kubernetes key: %s", key)
			}
		case "aliyun":
			switch key {
			case "region":
				cfg.Aliyun.Region = value
			case "credentialSource":
				cfg.Aliyun.CredentialSource = value
			case "accessKeyId":
				cfg.Aliyun.AccessKeyID = value
			case "accessKeySecret":
				cfg.Aliyun.AccessKeySecret = value
			case "casEndpoint":
				cfg.Aliyun.CASEndpoint = value
			case "cdnEndpoint":
				cfg.Aliyun.CDNEndpoint = value
			default:
				return Config{}, fmt.Errorf("unknown aliyun key: %s", key)
			}
		case "sync":
			switch key {
			case "maxRetries":
				n, err := strconv.Atoi(value)
				if err != nil {
					return Config{}, fmt.Errorf("invalid sync.maxRetries: %w", err)
				}
				cfg.Sync.MaxRetries = n
			case "retryBaseMillis":
				n, err := strconv.Atoi(value)
				if err != nil {
					return Config{}, fmt.Errorf("invalid sync.retryBaseMillis: %w", err)
				}
				cfg.Sync.RetryBaseMillis = n
			case "stateFile":
				cfg.Sync.StateFile = value
			default:
				return Config{}, fmt.Errorf("unknown sync key: %s", key)
			}
		default:
			return Config{}, fmt.Errorf("unknown section: %s", section)
		}
	}

	return cfg, nil
}
