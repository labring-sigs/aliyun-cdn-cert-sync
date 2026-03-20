package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTLSSecretFingerprint(t *testing.T) {
	certPEM, err := os.ReadFile(filepath.Join("..", "aliyun", "testdata", "cert.pem"))
	if err != nil {
		t.Fatalf("read sample cert: %v", err)
	}

	secret := TLSSecret{CertPEM: string(certPEM)}

	fingerprint, err := secret.Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint returned error: %v", err)
	}

	const want = "58:4A:C2:13:2E:1D:5A:7D:58:B4:18:BF:5B:97:B3:45:64:DE:26:C3"
	if fingerprint != want {
		t.Fatalf("expected fingerprint %q, got %q", want, fingerprint)
	}
}

func TestTLSSecretFingerprintRejectsInvalidPEM(t *testing.T) {
	secret := TLSSecret{CertPEM: "not a certificate"}

	_, err := secret.Fingerprint()
	if err == nil {
		t.Fatal("expected error for invalid cert pem")
	}
}
