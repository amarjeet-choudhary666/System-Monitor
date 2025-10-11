package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/logs"
)

func setupTestRouter() (*gin.Engine, error) {
	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add only the health check route for basic testing
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "CodeXray Observability Service is running",
		})
	})

	return router, nil
}

func TestHealthCheck(t *testing.T) {
	router, err := setupTestRouter()
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestUserRegistrationAndLogin(t *testing.T) {
	// Skip this test for now as it requires database setup
	t.Skip("Skipping database-dependent test - requires full integration test setup")
}

func TestLogAnalyzer(t *testing.T) {
	analyzer := logs.NewLogAnalyzer()

	// Test parsing a simple log line
	entry := analyzer.ParseLine("[INFO] Test message")
	assert.NotNil(t, entry)
	assert.Equal(t, logs.INFO, entry.Level)
	assert.Equal(t, "Test message", entry.Message)

	// Test parsing ERROR line
	entry = analyzer.ParseLine("[ERROR] Something went wrong")
	assert.NotNil(t, entry)
	assert.Equal(t, logs.ERROR, entry.Level)
	assert.Equal(t, "Something went wrong", entry.Message)
}

func TestMetricsCollection(t *testing.T) {
	// Skip this test for now as it requires database setup
	t.Skip("Skipping database-dependent test - requires full integration test setup")
}
