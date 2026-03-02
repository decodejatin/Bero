package api

import (
	"github.com/decodejatin/bero-backend/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Router sets up all API routes
type Router struct {
	echo           *echo.Echo
	authHandler    *AuthHandler
	jobHandler     *JobHandler
	profileHandler *ProfileHandler
	chatHandler    *ChatHandler
	addressHandler *AddressHandler
	ratingHandler  *RatingHandler
	authService    service.AuthService
}

// NewRouter creates a new router
func NewRouter(authHandler *AuthHandler, jobHandler *JobHandler, profileHandler *ProfileHandler, chatHandler *ChatHandler, addressHandler *AddressHandler, ratingHandler *RatingHandler, authService service.AuthService) *Router {
	return &Router{
		echo:           echo.New(),
		authHandler:    authHandler,
		jobHandler:     jobHandler,
		profileHandler: profileHandler,
		chatHandler:    chatHandler,
		addressHandler: addressHandler,
		ratingHandler:  ratingHandler,
		authService:    authService,
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

	// Job routes
	jobs := protected.Group("/jobs")
	jobs.POST("", r.jobHandler.CreateJob)
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

	return e
}

// Start starts the server
func (r *Router) Start(addr string) error {
	return r.echo.Start(addr)
}
