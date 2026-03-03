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
	"github.com/decodejatin/bero-backend/internal/matching"
	"github.com/decodejatin/bero-backend/internal/orchestrator"
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

	// Location services (geospatial infrastructure for Hungarian matching)
	locationRepo := repository.NewLocationRepository(db)
	locationService := service.NewLocationService(locationRepo)

	// Matching pre-processing (candidate filtering + weight matrix)
	candidateRepo := repository.NewCandidateRepository(db)
	matchingConfigRepo := repository.NewMatchingConfigRepository(db)
	matchingService := service.NewMatchingService(jobRepo, candidateRepo, matchingConfigRepo)

	// Matching engine (batched Hungarian algorithm)
	notificationHub := api.NewNotificationHub()
	dispatcher := matching.NewDispatcher(jobRepo, locationRepo, notificationHub.NotifyFunc(), nil)

	// Stability enforcement (wraps dispatch layer)
	stabilityRepo := repository.NewStabilityRepository(db)
	stabilityService := service.NewStabilityService(stabilityRepo, jobRepo)

	// Dynamic pricing engine (per-hexagon surge)
	pricingRepo := repository.NewPricingRepository(db)
	pricingService := service.NewPricingService(pricingRepo, jobRepo)

	// Unified Dispatch Orchestrator
	// Replaces BatchProcessor as the BatchQueue callback.
	// Coordinates: Pricing → Matching → Stability → Dispatch
	orch := orchestrator.NewOrchestrator(
		pricingService, matchingService, stabilityService,
		dispatcher, jobRepo, stabilityRepo, 30,
	)
	batchQueue := matching.NewBatchQueue(orch.ProcessBatch)
	orch.SetBatchQueue(batchQueue)
	dispatcher.SetBatchQueue(batchQueue)

	// Completion + mandatory ratings
	completionRatingRepo := repository.NewCompletionRatingRepository(db)
	completionService := service.NewCompletionService(jobRepo, userRepo, locationRepo, completionRatingRepo)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService)
	jobHandler := api.NewJobHandler(jobService)
	profileHandler := api.NewProfileHandler(profileService)
	chatHandler := api.NewChatHandler(chatService, authService)
	addressHandler := api.NewAddressHandler(addressService)
	ratingHandler := api.NewRatingHandler(ratingService)
	locationHandler := api.NewLocationHandler(locationService)
	matchingHandler := api.NewMatchingHandler(matchingService)
	matchingEngineHandler := api.NewMatchingEngineHandler(batchQueue, dispatcher)
	stabilityHandler := api.NewStabilityHandler(stabilityService)
	pricingHandler := api.NewPricingHandler(pricingService)
	orchestratorHandler := api.NewOrchestratorHandler(orch)
	completionHandler := api.NewCompletionHandler(completionService)

	// Initialize router
	router := api.NewRouter(authHandler, jobHandler, profileHandler, chatHandler, addressHandler, ratingHandler, locationHandler, matchingHandler, matchingEngineHandler, stabilityHandler, pricingHandler, orchestratorHandler, completionHandler, authService)
	e := router.Setup()

	// Start batch queue goroutine
	batchQueue.Start()
	defer batchQueue.Stop()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Server starting on %s", addr)
	log.Println("📚 API Documentation: http://localhost" + addr + "/health")

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx // available for future use
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("🛑 Shutting down...")
		cancel()
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
