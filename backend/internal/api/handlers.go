package api

import (
	"net/http"
	"strconv"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/alerts"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/auth"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/logs"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/metrics"
	"github.com/gin-gonic/gin"
)

// Handlers contains all API handlers
type Handlers struct {
	authService      *auth.Service
	logAnalyzer      *logs.LogAnalyzer
	metricsCollector *metrics.Collector
	alertService     *alerts.Service
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	authService *auth.Service,
	logAnalyzer *logs.LogAnalyzer,
	metricsCollector *metrics.Collector,
	alertService *alerts.Service,
) *Handlers {
	return &Handlers{
		authService:      authService,
		logAnalyzer:      logAnalyzer,
		metricsCollector: metricsCollector,
		alertService:     alertService,
	}
}


// Register handles user registration
func (h *Handlers) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles user authentication
func (h *Handlers) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// ValidateToken handles JWT token validation
func (h *Handlers) ValidateToken(c *gin.Context) {
	var req auth.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user":  user,
	})
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   newAccessToken,
		"message": "Token refreshed successfully",
	})
}

// Logout handles user logout (JWT is stateless, so this is just a success response)
func (h *Handlers) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// Log Analysis Handlers

// AnalyzeLogs handles log file analysis
func (h *Handlers) AnalyzeLogs(c *gin.Context) {
	filePath := c.Query("file")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter is required"})
		return
	}

	stats, err := h.logAnalyzer.ParseLogFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Log analysis completed",
		"stats":   stats,
	})
}

// Metrics Handlers

// GetCurrentMetrics returns current system metrics
func (h *Handlers) GetCurrentMetrics(c *gin.Context) {
	metrics, err := h.metricsCollector.GetCurrentMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Current metrics retrieved",
		"metrics": metrics,
	})
}

// GetMetricHistory returns historical metrics
func (h *Handlers) GetMetricHistory(c *gin.Context) {
	metricType := c.Param("type")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric type is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	history, err := h.metricsCollector.GetMetricHistory(metrics.MetricType(metricType), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metric history retrieved",
		"history": history,
	})
}

// Alert Handlers

// GetAlerts returns alerts with optional filtering
func (h *Handlers) GetAlerts(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	alertsList, err := h.alertService.GetAlerts(alerts.AlertStatus(status), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alerts retrieved",
		"alerts":  alertsList,
	})
}

// CreateAlert manually creates an alert (for testing)
func (h *Handlers) CreateAlert(c *gin.Context) {
	var req alerts.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert, err := h.alertService.CreateAlert(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert created",
		"alert":   alert,
	})
}

// ResolveAlert resolves an alert
func (h *Handlers) ResolveAlert(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert ID"})
		return
	}

	if err := h.alertService.ResolveAlert(uint(alertID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}

// Summary Handler

// GetSummary returns comprehensive system summary
func (h *Handlers) GetSummary(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// Get current metrics
	currentMetrics, err := h.metricsCollector.GetCurrentMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current metrics"})
		return
	}

	// Get alert summary
	alertSummary, err := h.alertService.GetAlertSummary(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get alert summary"})
		return
	}

	// Get metric summaries for last 10 readings
	cpuSummary, err := h.metricsCollector.GetMetricSummary(metrics.CPUUsage, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get CPU summary"})
		return
	}

	memorySummary, err := h.metricsCollector.GetMetricSummary(metrics.MemoryUsage, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get memory summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Summary retrieved",
		"summary": gin.H{
			"current_metrics": currentMetrics,
			"alerts":          alertSummary,
			"metric_averages": gin.H{
				"cpu":    cpuSummary,
				"memory": memorySummary,
			},
		},
	})
}

// Health check handler
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "CodeXray Observability Service is running",
	})
}
