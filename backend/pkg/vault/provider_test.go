package vault

import (
	"os"
	"testing"
)

// --- EnvProvider tests ---

func TestEnvProvider_ReadsEnv(t *testing.T) {
	os.Setenv("TEST_SECRET_KEY", "from_env")
	defer os.Unsetenv("TEST_SECRET_KEY")

	p := NewEnvProvider(nil)
	val, err := p.Get("test_secret_key")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if val != "from_env" {
		t.Errorf("expected 'from_env', got %q", val)
	}
}

func TestEnvProvider_FallsBackToDefaults(t *testing.T) {
	p := NewEnvProvider(map[string]string{
		"my_secret": "default_value",
	})
	val, err := p.Get("my_secret")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if val != "default_value" {
		t.Errorf("expected 'default_value', got %q", val)
	}
}

func TestEnvProvider_EnvOverridesDefault(t *testing.T) {
	os.Setenv("MY_SECRET", "env_override")
	defer os.Unsetenv("MY_SECRET")

	p := NewEnvProvider(map[string]string{
		"my_secret": "default_value",
	})
	val, err := p.Get("my_secret")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if val != "env_override" {
		t.Errorf("expected 'env_override', got %q", val)
	}
}

func TestEnvProvider_KeyNotFound(t *testing.T) {
	p := NewEnvProvider(nil)
	_, err := p.Get("nonexistent_key")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

// --- MemoryProvider tests ---

func TestMemoryProvider_StoreAndRetrieve(t *testing.T) {
	p := NewMemoryProvider(map[string]string{
		"db_password": "secure123",
		"jwt_secret":  "my-jwt-key",
	})

	val, err := p.Get("db_password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secure123" {
		t.Errorf("expected 'secure123', got %q", val)
	}
}

func TestMemoryProvider_Set(t *testing.T) {
	p := NewMemoryProvider(nil)
	p.Set("api_key", "new_key_value")

	val, err := p.Get("api_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "new_key_value" {
		t.Errorf("expected 'new_key_value', got %q", val)
	}
}

func TestMemoryProvider_KeyNotFound(t *testing.T) {
	p := NewMemoryProvider(nil)
	_, err := p.Get("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

// --- Factory tests ---

func TestNewProvider_DefaultIsEnv(t *testing.T) {
	p, err := NewProvider("env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer p.Close()

	// Should have dev defaults
	val, err := p.Get("db_password")
	if err != nil {
		t.Fatalf("expected default db_password, got error: %v", err)
	}
	if val != "password" {
		t.Errorf("expected default 'password', got %q", val)
	}
}

func TestNewProvider_MemoryIsEmpty(t *testing.T) {
	p, err := NewProvider("memory")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer p.Close()

	_, err = p.Get("anything")
	if err == nil {
		t.Error("expected error from empty memory provider")
	}
}

func TestNewProvider_VaultFailsWithoutAddr(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")

	_, err := NewProvider("vault")
	if err == nil {
		t.Error("expected error when VAULT_ADDR is missing")
	}
}

// --- Helper tests ---

func TestEnvKeyFromSecret(t *testing.T) {
	cases := map[string]string{
		"db_password": "DB_PASSWORD",
		"jwt_secret":  "JWT_SECRET",
		"ALREADY_UP":  "ALREADY_UP",
		"redis_url":   "REDIS_URL",
	}
	for input, want := range cases {
		if got := envKeyFromSecret(input); got != want {
			t.Errorf("envKeyFromSecret(%q) = %q, want %q", input, got, want)
		}
	}
}
