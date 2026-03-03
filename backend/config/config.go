package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerPort string
	GRPCPort   string // internal gRPC port (default: 50051)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Infrastructure
	NATSUrl      string // NATS server URL (default: nats://localhost:4222)
	RedisUrl     string // Redis server URL (default: redis://localhost:6379)
	EnableShadow bool   // Enable shadow matching mode
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		GRPCPort:     getEnv("GRPC_PORT", "50051"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "password"),
		DBName:       getEnv("DB_NAME", "bero"),
		NATSUrl:      getEnv("NATS_URL", "nats://localhost:4222"),
		RedisUrl:     getEnv("REDIS_URL", "redis://localhost:6379"),
		EnableShadow: getBoolEnv("ENABLE_SHADOW_MATCHING", false),
	}
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
