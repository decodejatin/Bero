package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/decodejatin/bero-backend/config"
	jobpb "github.com/decodejatin/bero-backend/gen/pb/job"
	matcherpb "github.com/decodejatin/bero-backend/gen/pb/matcher"
	userpb "github.com/decodejatin/bero-backend/gen/pb/user"
	"github.com/decodejatin/bero-backend/internal/api"
	grpcserver "github.com/decodejatin/bero-backend/internal/grpc/server"
	"github.com/decodejatin/bero-backend/internal/matching"
	"github.com/decodejatin/bero-backend/internal/matchmaker"
	"github.com/decodejatin/bero-backend/internal/orchestrator"
	"github.com/decodejatin/bero-backend/internal/repository"
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/decodejatin/bero-backend/pkg/database"
	"github.com/decodejatin/bero-backend/pkg/eventbus"
	"google.golang.org/grpc"
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

	// Initialize infrastructure
	bus := eventbus.NewInMemoryEventBus()
	log.Printf("📡 Event bus initialized (in-memory, NATS_URL=%s for future swap)", cfg.NATSUrl)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)
	jobRepo := repository.NewJobRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// Initialize services
	jwtCfg := service.JWTConfig{
		SecretKey:       cfg.JWTSecret, // loaded via SecretProvider (Vault in production)
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

	// Matchmaker service
	matchmakerRepo := repository.NewMatchmakerRepository(db)
	matchCfg := matchmaker.DefaultConfig()
	matchCfg.EnableShadowMode = cfg.EnableShadow
	matchmakerService := service.NewMatchmakerService(matchmakerRepo, jobRepo, matchCfg)

	// Set up event subscriptions
	bus.Subscribe(eventbus.EventOrderPlaced, func(e eventbus.Event) {
		log.Printf("[event] Order placed: %s — triggering matchmaker", e.ID)
		matchmakerService.Trigger()
	})
	_ = bus

	// ── gRPC Server ───────────────────────────────────────────────────────────
	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	grpcSrv := grpc.NewServer()

	// Register all gRPC services
	userpb.RegisterUserServiceServer(grpcSrv, grpcserver.NewUserServer(userRepo, matchmakerRepo))
	jobpb.RegisterJobServiceServer(grpcSrv, grpcserver.NewJobServer(jobRepo))
	// MatcherServer needs direct access to the engine — extract it via a helper
	matcherpb.RegisterMatcherServiceServer(grpcSrv, grpcserver.NewMatcherServer(
		matchmakerService.GetEngine(),
	))

	go func() {
		log.Printf("⚡ gRPC server listening on %s (binary Protobuf)", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()
	// ─────────────────────────────────────────────────────────────────────────

	// Initialize HTTP handlers
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
	log.Printf("🚀 HTTP server starting on %s", addr)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx // available for future use
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("🛑 Shutting down...")
		cancel()
		grpcSrv.GracefulStop()
		if err := e.Close(); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	if err := e.Start(addr); err != nil {
		log.Printf("Server stopped: %v", err)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultVal
}
