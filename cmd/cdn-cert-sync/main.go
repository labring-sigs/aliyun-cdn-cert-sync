package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/labring-sigs/aliyun-cdn-cert-sync/internal/aliyun"
	"github.com/labring-sigs/aliyun-cdn-cert-sync/internal/config"
	"github.com/labring-sigs/aliyun-cdn-cert-sync/internal/k8s"
	"github.com/labring-sigs/aliyun-cdn-cert-sync/internal/sync"
)

func main() {
	configPath := flag.String("config", "./configs/config.example.yaml", "path to config file")
	adapterMode := flag.String("adapter-mode", "memory", "adapter mode: memory|api")
	inCluster := flag.Bool("in-cluster", true, "use in-cluster kubernetes client")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (used when --in-cluster=false)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("validate config: %v", err)
	}

	var (
		secretSource     k8s.SecretSource
		certificateStore aliyun.CertificateStore
		cdnBinder        aliyun.CDNBinder
		stateStore       sync.StateStore
	)

	stateStore, err = sync.NewFileStateStore(cfg.Sync.StateFile)
	if err != nil {
		log.Fatalf("init state store: %v", err)
	}

	switch *adapterMode {
	case "memory":
		secretSource = k8s.NewMemorySecretSource(cfg.Kubernetes.SecretNamespace, cfg.Kubernetes.SecretName)
		certificateStore = aliyun.NewMemoryCertificateStore()
		cdnBinder = aliyun.NewMemoryCDNBinder()
	case "api":
		k8sClient, err := k8s.NewClient(k8s.ClientConfig{
			InCluster:  *inCluster,
			Kubeconfig: *kubeconfig,
		})
		if err != nil {
			log.Fatalf("init kubernetes client: %v", err)
		}

		aliyunClient, err := aliyun.NewAPIClient(aliyun.APIClientConfig{
			Region:           cfg.Aliyun.Region,
			CredentialSource: cfg.Aliyun.CredentialSource,
			AccessKeyID:      cfg.Aliyun.AccessKeyID,
			AccessKeySecret:  cfg.Aliyun.AccessKeySecret,
			CASEndpoint:      cfg.Aliyun.CASEndpoint,
			CDNEndpoint:      cfg.Aliyun.CDNEndpoint,
			ResourceGroupID:  cfg.Aliyun.ResourceGroupID,
		})
		if err != nil {
			log.Fatalf("init aliyun client: %v", err)
		}

		secretSource = k8s.NewAPISecretSource(k8sClient)
		certificateStore = aliyun.NewCASCertificateStore(aliyunClient)
		cdnBinder = aliyun.NewAPICDNBinder(aliyunClient)
	default:
		log.Fatalf("unsupported adapter mode %q (expected memory|api)", *adapterMode)
	}

	reconciler := sync.NewReconciler(
		cfg,
		secretSource,
		certificateStore,
		cdnBinder,
		stateStore,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := reconciler.RunOnce(ctx)
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	log.Printf(
		"sync complete: uploaded=%t cert_id=%s domains_updated=%d domain_failures=%d retries=%d",
		report.Uploaded,
		report.CertificateID,
		report.DomainsUpdated,
		report.DomainFailures,
		report.Retries,
	)
}
