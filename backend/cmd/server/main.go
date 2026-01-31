package main

import (
	"fmt"
	"log"
	"time"

	"github.com/decodejatin/bero-backend/config"
	"github.com/decodejatin/bero-backend/internal/api"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/decodejatin/bero-backend/pkg/database"
)

// @title Bero API
// @version 1.0
// @description Backend API for Bero - Home Services Marketplace
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// Initialize services
	jwtCfg := service.JWTConfig{
		SecretKey:       getEnvOrDefault("JWT_SECRET", "super-secret-key-change-in-production"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	authService := service.NewAuthService(authRepo, userRepo, jwtCfg)
	jobService := service.NewJobService(jobRepo)
	profileService := service.NewProfileService(userRepo)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	jobHandler := api.NewJobHandler(jobService)
	profileHandler := api.NewProfileHandler(profileService)

	// Initialize router
	router := api.NewRouter(authHandler, jobHandler, profileHandler, authService)
	e := router.Setup()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on %s", addr)
	log.Println("📚 API Documentation: http://localhost" + addr + "/health")
	if err := e.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	_ = key // TODO: Use os.Getenv(key) when integrating real env vars
	return defaultVal
}
