package config

import (
	"log"
	"os"
	"strings"

	"github.com/decodejatin/bero-backend/pkg/vault"
)

type Config struct {
	ServerPort string
	GRPCPort   string // internal gRPC port (default: 50051)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Secrets loaded via SecretProvider (Vault in production)
	JWTSecret string // JWT signing key — NEVER hardcode in production

	// Infrastructure
	NATSUrl      string // NATS server URL (default: nats://localhost:4222)
	RedisUrl     string // Redis server URL (default: redis://localhost:6379)
	EnableShadow bool   // Enable shadow matching mode

	// Secret management
	SecretProvider string // "env" (default) or "vault"
	VaultAddr      string // Vault server URL (when SecretProvider=vault)
	VaultToken     string // Vault auth token (bootstrap only — use AppRole in production)
	VaultPath      string // Vault KV v2 path prefix (default: secret/data/bero)
}

func LoadConfig() *Config {
	cfg := &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		GRPCPort:     getEnv("GRPC_PORT", "50051"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBName:       getEnv("DB_NAME", "bero"),
		NATSUrl:      getEnv("NATS_URL", "nats://localhost:4222"),
		RedisUrl:     getEnv("REDIS_URL", "redis://localhost:6379"),
		EnableShadow: getBoolEnv("ENABLE_SHADOW_MATCHING", false),

		// Vault config (only needed when SECRET_PROVIDER=vault)
		SecretProvider: getEnv("SECRET_PROVIDER", "env"),
		VaultAddr:      getEnv("VAULT_ADDR", ""),
		VaultToken:     getEnv("VAULT_TOKEN", ""),
		VaultPath:      getEnv("VAULT_PATH", "secret/data/bero"),
	}

	// Load secrets via the appropriate provider
	secrets, err := vault.NewProvider(cfg.SecretProvider)
	if err != nil {
		log.Fatalf("❌ Failed to initialize secret provider (%s): %v", cfg.SecretProvider, err)
	}

	// Read sensitive values through the provider
	cfg.DBPassword = secrets.MustGet("db_password")
	cfg.JWTSecret = secrets.MustGet("jwt_secret")

	log.Printf("🔐 Secrets loaded via %s provider", cfg.SecretProvider)
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.EqualFold(value, "true") || value == "1"
	}
	return fallback
}
