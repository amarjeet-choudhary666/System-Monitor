package api

import (
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, handlers *Handlers, authService *auth.Service) {
	// Add middleware
	router.Use(CORSMiddleware())
	router.Use(LoggingMiddleware())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Authentication routes (public)
	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/register", handlers.Register)
		authRoutes.POST("/login", handlers.Login)
		authRoutes.POST("/validate", handlers.ValidateToken)
		authRoutes.POST("/refresh", handlers.RefreshToken)
	}

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(AuthMiddleware(authService))
	{
		// Auth routes
		protected.POST("/auth/logout", handlers.Logout)

		// Log analysis routes
		logRoutes := protected.Group("/logs")
		{
			logRoutes.GET("/analyze", handlers.AnalyzeLogs)
		}

		// Metrics routes
		metricsRoutes := protected.Group("/metrics")
		{
			metricsRoutes.GET("/current", handlers.GetCurrentMetrics)
			metricsRoutes.GET("/history/:type", handlers.GetMetricHistory)
		}

		// Alert routes
		alertRoutes := protected.Group("/alerts")
		{
			alertRoutes.GET("", handlers.GetAlerts)
			alertRoutes.POST("", handlers.CreateAlert)
			alertRoutes.PUT("/:id/resolve", handlers.ResolveAlert)
		}

		// Summary route
		protected.GET("/summary", handlers.GetSummary)
	}
}
