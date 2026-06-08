package api

import (
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Router sets up all API routes
type Router struct {
	echo                  *echo.Echo
	authHandler           *AuthHandler
	jobHandler            *JobHandler
	profileHandler        *ProfileHandler
	chatHandler           *ChatHandler
	addressHandler        *AddressHandler
	ratingHandler         *RatingHandler
	locationHandler       *LocationHandler
	matchingHandler       *MatchingHandler
	matchingEngineHandler *MatchingEngineHandler
	stabilityHandler      *StabilityHandler
	pricingHandler        *PricingHandler
	orchestratorHandler   *OrchestratorHandler
	completionHandler     *CompletionHandler
	legalHandler          *LegalHandler
	authService           service.AuthService
	legalService          *service.LegalService
}

// NewRouter creates a new router
func NewRouter(authHandler *AuthHandler, jobHandler *JobHandler, profileHandler *ProfileHandler, chatHandler *ChatHandler, addressHandler *AddressHandler, ratingHandler *RatingHandler, locationHandler *LocationHandler, matchingHandler *MatchingHandler, matchingEngineHandler *MatchingEngineHandler, stabilityHandler *StabilityHandler, pricingHandler *PricingHandler, orchestratorHandler *OrchestratorHandler, completionHandler *CompletionHandler, legalHandler *LegalHandler, legalService *service.LegalService, authService service.AuthService) *Router {
	return &Router{
		echo:                  echo.New(),
		authHandler:           authHandler,
		jobHandler:            jobHandler,
		profileHandler:        profileHandler,
		chatHandler:           chatHandler,
		addressHandler:        addressHandler,
		ratingHandler:         ratingHandler,
		locationHandler:       locationHandler,
		matchingHandler:       matchingHandler,
		matchingEngineHandler: matchingEngineHandler,
		stabilityHandler:      stabilityHandler,
		pricingHandler:        pricingHandler,
		orchestratorHandler:   orchestratorHandler,
		completionHandler:     completionHandler,
		legalHandler:          legalHandler,
		authService:           authService,
		legalService:          legalService,
	}
}

// Setup configures all routes and middleware
func (r *Router) Setup() *echo.Echo {
	e := r.echo

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "bero-backend",
		})
	})

	// API v1
	v1 := e.Group("/api/v1")

	// Public auth routes
	auth := v1.Group("/auth")
	auth.POST("/send-otp", r.authHandler.SendOtp)
	auth.POST("/verify-otp", r.authHandler.VerifyOtp)
	auth.POST("/refresh", r.authHandler.RefreshToken)

	// Protected routes
	protected := v1.Group("")
	protected.Use(AuthMiddleware(r.authService))

	// Protected auth routes
	protectedAuth := protected.Group("/auth")
	protectedAuth.POST("/logout", r.authHandler.Logout)
	protectedAuth.GET("/me", r.authHandler.Me)

	// Profile routes
	profile := protected.Group("/profile")
	profile.GET("", r.profileHandler.GetProfile)
	profile.PUT("", r.profileHandler.UpdateProfile)
	profile.POST("/user-type", r.profileHandler.SetUserType)
	profile.GET("/stats", r.profileHandler.GetUserStats)
	profile.GET("/:id", r.profileHandler.GetProfileById)
	profile.PUT("/worker/skills", r.profileHandler.UpdateWorkerSkills)
	profile.GET("/ratings", r.profileHandler.GetMyRatings)

	// Job routes (with legal compliance check on job creation)
	jobs := protected.Group("/jobs")
	legalCheck := LegalComplianceMiddleware(r.legalService)
	jobs.POST("", r.jobHandler.CreateJob, legalCheck)
	jobs.GET("", r.jobHandler.GetAvailableJobs)
	jobs.GET("/my", r.jobHandler.GetMyJobs)
	jobs.GET("/:id", r.jobHandler.GetJob)
	jobs.POST("/:id/accept", r.jobHandler.AcceptJob)
	jobs.POST("/:id/start", r.jobHandler.StartJob)
	jobs.POST("/:id/complete", r.jobHandler.CompleteJob)
	jobs.POST("/:id/cancel", r.jobHandler.CancelJob)
	jobs.POST("/:id/confirm", r.jobHandler.ConfirmCompletion)
	jobs.POST("/:id/rate", r.ratingHandler.SubmitRating)
	jobs.GET("/:id/ratings", r.ratingHandler.GetJobRatings)

	// Address routes
	addresses := protected.Group("/addresses")
	addresses.GET("", r.addressHandler.GetAddresses)
	addresses.POST("", r.addressHandler.CreateAddress)
	addresses.PUT("/:id", r.addressHandler.UpdateAddress)
	addresses.DELETE("/:id", r.addressHandler.DeleteAddress)

	// Chat routes (REST — protected)
	chat := protected.Group("/chat")
	chat.GET("/conversations", r.chatHandler.GetConversations)
	chat.POST("/conversations", r.chatHandler.CreateOrGetConversation)
	chat.GET("/conversations/:id/messages", r.chatHandler.GetMessages)
	chat.POST("/conversations/:id/messages", r.chatHandler.SendMessage)
	chat.PUT("/conversations/:id/read", r.chatHandler.MarkAsRead)
	chat.GET("/unread", r.chatHandler.GetUnreadCount)

	// Chat WebSocket (auth via query param, not middleware)
	v1.GET("/chat/ws", r.chatHandler.HandleWebSocket)

	// Location & geospatial routes
	workers := protected.Group("/workers")
	workers.PUT("/location", r.locationHandler.UpdateWorkerLocation)
	workers.GET("/nearby", r.locationHandler.GetNearbyWorkers)
	workers.PUT("/availability", r.locationHandler.SetAvailability, LegalComplianceMiddleware(r.legalService))
	workers.DELETE("/location", r.locationHandler.GoOffline)

	// Matching pre-processing routes (feeds into Hungarian algorithm)
	matching := protected.Group("/matching")
	matching.GET("/candidates/:jobId", r.matchingHandler.GetCandidates)
	matching.GET("/weights", r.matchingHandler.GetWeights)
	matching.PUT("/weights", r.matchingHandler.UpdateWeights)
	matching.POST("/matrix", r.matchingHandler.BuildMatrix)

	// Matching engine routes (batched Hungarian)
	matching.POST("/enqueue", r.matchingEngineHandler.EnqueueJob)
	matching.GET("/queue/status", r.matchingEngineHandler.GetQueueStatus)
	matching.POST("/batch/trigger", r.matchingEngineHandler.TriggerBatch)
	matching.POST("/decline/:jobId", r.matchingEngineHandler.DeclineJob)

	// Stability enforcement routes
	stability := protected.Group("/stability")
	stability.GET("/stats", r.stabilityHandler.GetStats)
	stability.GET("/config", r.stabilityHandler.GetConfig)
	stability.PUT("/config", r.stabilityHandler.UpdateConfig)
	stability.GET("/user/:id/status", r.stabilityHandler.GetUserStatus)
	stability.POST("/reassign", r.stabilityHandler.CheckReassign)
	stability.POST("/utility", r.stabilityHandler.ComputeUtility)

	// Dynamic pricing routes
	pricing := protected.Group("/pricing")
	pricing.GET("/surge", r.pricingHandler.GetSurge)
	pricing.GET("/quote/:jobId", r.pricingHandler.GetPriceQuote)
	pricing.POST("/apply/:jobId", r.pricingHandler.ApplySurge)
	pricing.GET("/config", r.pricingHandler.GetConfig)
	pricing.PUT("/config", r.pricingHandler.UpdateConfig)
	pricing.GET("/history", r.pricingHandler.GetSurgeHistory)

	// Pipeline orchestrator routes
	pipeline := protected.Group("/pipeline")
	pipeline.POST("/submit", r.orchestratorHandler.SubmitJob)
	pipeline.GET("/status", r.orchestratorHandler.GetStatus)

	// Completion + rating routes
	protected.POST("/jobs/:id/complete-by-worker", r.completionHandler.WorkerComplete)
	protected.POST("/jobs/:id/confirm-by-client", r.completionHandler.ClientConfirm)
	protected.POST("/jobs/:id/rate", r.completionHandler.SubmitRating)
	protected.GET("/jobs/:id/completion-status", r.completionHandler.GetCompletionStatus)
	completion := protected.Group("/completion")
	completion.GET("/blocked/worker/:id", r.completionHandler.CheckWorkerBlocked)
	completion.GET("/blocked/client/:id", r.completionHandler.CheckClientBlocked)

	// Legal documents routes (public — document list)
	v1.GET("/legal/documents", r.legalHandler.GetDocuments)
	v1.GET("/legal/documents/:slug", r.legalHandler.GetDocument)

	// Legal routes (authenticated)
	legal := protected.Group("/legal")
	legal.POST("/accept", r.legalHandler.AcceptDocuments)
	legal.POST("/accept-worker-policy", r.legalHandler.AcceptWorkerPolicy)
	legal.GET("/compliance", r.legalHandler.GetComplianceStatus)

	// Admin legal routes (authenticated — add admin middleware in production)
	admin := protected.Group("/admin")
	admin.PUT("/legal/documents/:slug", r.legalHandler.AdminUpdateDocument)
	admin.POST("/legal/documents/:slug/force-reaccept", r.legalHandler.AdminForceReAccept)
	admin.GET("/legal/acceptance-logs", r.legalHandler.AdminGetAcceptanceLogs)

	return e
}

// Start starts the server
func (r *Router) Start(addr string) error {
	return r.echo.Start(addr)
}
