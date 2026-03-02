package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/decodejatin/bero-backend/config"
	"github.com/decodejatin/bero-backend/internal/api"
	"github.com/decodejatin/bero-backend/internal/matchmaker"
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
	chatRepo := repository.NewChatRepository(db)
	matchmakerRepo := repository.NewMatchmakerRepository(db)

	// Initialize services
	jwtCfg := service.JWTConfig{
		SecretKey:       getEnvOrDefault("JWT_SECRET", "super-secret-key-change-in-production"),
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	authService := service.NewAuthService(authRepo, userRepo, jwtCfg)
	jobService := service.NewJobService(jobRepo, userRepo)
	profileService := service.NewProfileService(userRepo, jobRepo)
	chatService := service.NewChatService(chatRepo, userRepo)

	// Address & Rating services
	addressRepo := repository.NewAddressRepository(db)
	addressService := service.NewAddressService(addressRepo)
	ratingService := service.NewRatingService(jobRepo, userRepo)

	// Matchmaker service
	matchCfg := matchmaker.DefaultConfig()
	matchmakerService := service.NewMatchmakerService(matchmakerRepo, jobRepo, matchCfg)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	jobHandler := api.NewJobHandler(jobService)
	profileHandler := api.NewProfileHandler(profileService)
	chatHandler := api.NewChatHandler(chatService, authService)
	addressHandler := api.NewAddressHandler(addressService)
	ratingHandler := api.NewRatingHandler(ratingService)
	matchmakerHandler := api.NewMatchmakerHandler(matchmakerService)

	// Initialize router
	router := api.NewRouter(authHandler, jobHandler, profileHandler, chatHandler, addressHandler, ratingHandler, matchmakerHandler, authService)
	e := router.Setup()

	// Start matchmaker engine with graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	matchmakerService.StartEngine(ctx)
	log.Println("🧠 Matchmaker engine started")

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on %s", addr)
	log.Println("📚 API Documentation: http://localhost" + addr + "/health")

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("🛑 Shutting down...")
		cancel() // Stop matchmaker engine
		if err := e.Close(); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := e.Start(addr); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	_ = key // TODO: Use os.Getenv(key) when integrating real env vars
	return defaultVal
}
