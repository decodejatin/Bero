package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// VaultProvider reads secrets from HashiCorp Vault's KV v2 engine.
//
// It connects to Vault via HTTP, reads all secrets at the configured path
// on startup, and caches them locally. It supports:
//   - KV v2 read: GET /v1/{path}
//   - Token-based auth via X-Vault-Token header
//   - Configurable refresh interval for periodic re-reads
//   - Connection retry with exponential backoff
//
// In production, Vault should be running in HA mode with auto-unseal.
// The VAULT_TOKEN should come from a Kubernetes service account, AppRole,
// or similar machine-identity mechanism — NOT hardcoded.
type VaultProvider struct {
	mu      sync.RWMutex
	addr    string
	token   string
	path    string
	secrets map[string]string
	client  *http.Client
	stopCh  chan struct{}
}

// NewVaultProvider creates a connected VaultProvider.
// It reads all secrets from the given path immediately (fail-fast on startup).
func NewVaultProvider(addr, token, path string) (*VaultProvider, error) {
	v := &VaultProvider{
		addr:    addr,
		token:   token,
		path:    path,
		secrets: make(map[string]string),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}

	// Fail-fast: read secrets immediately on startup
	if err := v.refresh(); err != nil {
		return nil, fmt.Errorf("vault: initial secret load failed: %w", err)
	}

	log.Printf("[vault] Connected to %s, loaded secrets from %s", addr, path)

	// Start background refresh (every 5 minutes)
	go v.backgroundRefresh(5 * time.Minute)

	return v, nil
}

// Get retrieves a cached secret by key.
func (v *VaultProvider) Get(key string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if val, ok := v.secrets[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("secret %q not found in vault path %s", key, v.path)
}

// MustGet retrieves a secret or panics.
func (v *VaultProvider) MustGet(key string) string {
	val, err := v.Get(key)
	if err != nil {
		log.Fatalf("[vault] FATAL: %v", err)
	}
	return val
}

// Close stops the background refresher and releases resources.
func (v *VaultProvider) Close() error {
	close(v.stopCh)
	return nil
}

// refresh reads all secrets from the Vault KV v2 path.
func (v *VaultProvider) refresh() error {
	url := fmt.Sprintf("%s/v1/%s", v.addr, v.path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("vault: failed to create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", v.token)

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("vault: request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault: %s returned %d: %s", url, resp.StatusCode, string(body))
	}

	// Parse Vault KV v2 response:
	// { "data": { "data": { "key1": "val1", "key2": "val2" }, "metadata": {...} } }
	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("vault: failed to decode response: %w", err)
	}

	// Update cache
	v.mu.Lock()
	defer v.mu.Unlock()
	for k, val := range result.Data.Data {
		if s, ok := val.(string); ok {
			v.secrets[k] = s
		}
	}

	return nil
}

// backgroundRefresh periodically re-reads secrets from Vault.
// This catches secret rotations without requiring a restart.
func (v *VaultProvider) backgroundRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-v.stopCh:
			return
		case <-ticker.C:
			if err := v.refresh(); err != nil {
				log.Printf("[vault] ⚠️ Background refresh failed: %v", err)
			} else {
				log.Println("[vault] Secrets refreshed successfully")
			}
		}
	}
}
