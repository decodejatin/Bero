// Package vault provides a SecretProvider abstraction for retrieving secrets.
//
// In development, secrets are read from environment variables (EnvProvider).
// In production, secrets are read from HashiCorp Vault (VaultProvider).
// In tests, secrets come from an in-memory map (MemoryProvider).
//
// Usage:
//
//	provider, _ := vault.NewProvider("env")             // development
//	provider, _ := vault.NewProvider("vault")           // production (needs VAULT_ADDR, VAULT_TOKEN)
//	provider     = vault.NewMemoryProvider(map[...]{})  // unit tests
//
//	password := provider.MustGet("db_password")
package vault

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// SecretProvider is the interface for retrieving secrets.
// All implementations must be safe for concurrent use.
type SecretProvider interface {
	// Get retrieves a secret by key. Returns an error if the key is not found.
	Get(key string) (string, error)

	// MustGet retrieves a secret by key or panics. Use for startup-critical secrets.
	MustGet(key string) string

	// Close releases any resources (e.g., Vault connections).
	Close() error
}

// NewProvider creates a SecretProvider based on the provider type.
//   - "env"    → reads from OS environment variables (default, for development)
//   - "vault"  → reads from HashiCorp Vault KV v2 (for production)
//   - "memory" → returns an empty MemoryProvider (for tests)
func NewProvider(providerType string) (SecretProvider, error) {
	switch providerType {
	case "vault":
		addr := os.Getenv("VAULT_ADDR")
		token := os.Getenv("VAULT_TOKEN")
		path := os.Getenv("VAULT_PATH")
		if addr == "" {
			return nil, fmt.Errorf("VAULT_ADDR is required when SECRET_PROVIDER=vault")
		}
		if token == "" {
			return nil, fmt.Errorf("VAULT_TOKEN is required when SECRET_PROVIDER=vault")
		}
		if path == "" {
			path = "secret/data/bero"
		}
		return NewVaultProvider(addr, token, path)

	case "memory":
		return NewMemoryProvider(nil), nil

	default: // "env"
		return NewEnvProvider(defaultSecrets()), nil
	}
}

// --- EnvProvider -----------------------------------------------------------

// EnvProvider reads secrets from OS environment variables.
// Falls back to a defaults map for development convenience.
// NEVER use the defaults in production — they are only for local development.
type EnvProvider struct {
	defaults map[string]string
}

// NewEnvProvider creates an EnvProvider with the given fallback defaults.
func NewEnvProvider(defaults map[string]string) *EnvProvider {
	if defaults == nil {
		defaults = make(map[string]string)
	}
	return &EnvProvider{defaults: defaults}
}

// Get reads a secret from the environment, falling back to defaults.
func (p *EnvProvider) Get(key string) (string, error) {
	// Check environment first (uppercase convention: db_password → DB_PASSWORD)
	envKey := envKeyFromSecret(key)
	if v, ok := os.LookupEnv(envKey); ok {
		return v, nil
	}
	// Check defaults
	if v, ok := p.defaults[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("secret %q not found in environment or defaults", key)
}

// MustGet reads a secret or panics.
func (p *EnvProvider) MustGet(key string) string {
	v, err := p.Get(key)
	if err != nil {
		log.Fatalf("[vault] FATAL: %v", err)
	}
	return v
}

// Close is a no-op for EnvProvider.
func (p *EnvProvider) Close() error { return nil }

// --- MemoryProvider --------------------------------------------------------

// MemoryProvider stores secrets in-memory. Used in unit tests.
type MemoryProvider struct {
	mu      sync.RWMutex
	secrets map[string]string
}

// NewMemoryProvider creates a MemoryProvider pre-loaded with the given secrets.
func NewMemoryProvider(secrets map[string]string) *MemoryProvider {
	if secrets == nil {
		secrets = make(map[string]string)
	}
	return &MemoryProvider{secrets: secrets}
}

// Set adds or updates a secret in memory.
func (p *MemoryProvider) Set(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.secrets[key] = value
}

// Get retrieves a secret from memory.
func (p *MemoryProvider) Get(key string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if v, ok := p.secrets[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("secret %q not found in memory provider", key)
}

// MustGet retrieves a secret or panics.
func (p *MemoryProvider) MustGet(key string) string {
	v, err := p.Get(key)
	if err != nil {
		log.Fatalf("[vault] FATAL: %v", err)
	}
	return v
}

// Close is a no-op for MemoryProvider.
func (p *MemoryProvider) Close() error { return nil }

// --- Helpers ---------------------------------------------------------------

// envKeyFromSecret converts a secret key to its environment variable name.
// Convention: lowercase_snake → UPPERCASE_SNAKE
func envKeyFromSecret(key string) string {
	result := make([]byte, len(key))
	for i, c := range key {
		if c >= 'a' && c <= 'z' {
			result[i] = byte(c - 32) // to uppercase
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

// defaultSecrets returns development defaults for all known secrets.
// ⚠️ These MUST NOT be used in production.
func defaultSecrets() map[string]string {
	return map[string]string{
		"db_password": "password",
		"jwt_secret":  "super-secret-key-change-in-production",
	}
}
