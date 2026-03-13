package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type StateStore interface {
	GetCertIDByFingerprint(fingerprint string) (string, bool, error)
	SetCertIDByFingerprint(fingerprint, certID string) error
}

type FileStateStore struct {
	path string
	mu   sync.Mutex
}

type statePayload struct {
	FingerprintToCertID map[string]string `json:"fingerprintToCertId"`
}

func NewFileStateStore(path string) (*FileStateStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("state path is required")
	}
	return &FileStateStore{path: path}, nil
}

func (s *FileStateStore) GetCertIDByFingerprint(fingerprint string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.read()
	if err != nil {
		return "", false, err
	}
	id, ok := state.FingerprintToCertID[fingerprint]
	return id, ok, nil
}

func (s *FileStateStore) SetCertIDByFingerprint(fingerprint, certID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.read()
	if err != nil {
		return err
	}
	state.FingerprintToCertID[fingerprint] = certID
	return s.write(state)
}

func (s *FileStateStore) read() (statePayload, error) {
	if _, err := os.Stat(s.path); errors.Is(err, os.ErrNotExist) {
		return statePayload{FingerprintToCertID: map[string]string{}}, nil
	}

	raw, err := os.ReadFile(s.path)
	if err != nil {
		return statePayload{}, fmt.Errorf("read state file: %w", err)
	}
	if len(raw) == 0 {
		return statePayload{FingerprintToCertID: map[string]string{}}, nil
	}

	var payload statePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return statePayload{}, fmt.Errorf("unmarshal state: %w", err)
	}
	if payload.FingerprintToCertID == nil {
		payload.FingerprintToCertID = map[string]string{}
	}
	return payload, nil
}

func (s *FileStateStore) write(state statePayload) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(s.path, raw, 0o600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}
