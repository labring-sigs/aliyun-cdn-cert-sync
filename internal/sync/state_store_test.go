package sync

import (
	"path/filepath"
	"testing"
)

func TestFileStateStoreRoundTrip(t *testing.T) {
	store, err := NewFileStateStore(filepath.Join(t.TempDir(), "certmap.json"))
	if err != nil {
		t.Fatalf("NewFileStateStore returned error: %v", err)
	}

	_, ok, err := store.GetCertIDByFingerprint("fp-a")
	if err != nil {
		t.Fatalf("GetCertIDByFingerprint returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected missing fingerprint")
	}

	if err := store.SetCertIDByFingerprint("fp-a", "cert-1"); err != nil {
		t.Fatalf("SetCertIDByFingerprint returned error: %v", err)
	}

	id, ok, err := store.GetCertIDByFingerprint("fp-a")
	if err != nil {
		t.Fatalf("GetCertIDByFingerprint returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected fingerprint mapping to exist")
	}
	if id != "cert-1" {
		t.Fatalf("expected cert-1, got %s", id)
	}
}
