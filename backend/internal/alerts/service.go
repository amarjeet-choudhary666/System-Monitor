package alerts

import (
	"fmt"
	"log"
	"time"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/metrics"
	"gorm.io/gorm"
)

// Service handles alert operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new alert service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CheckThresholds checks if current metrics exceed thresholds and creates alerts
func (s *Service) CheckThresholds(currentMetrics *metrics.SystemMetrics) error {
	// Get all enabled thresholds
	var thresholds []metrics.MetricThreshold
	if err := s.db.Where("enabled = ?", true).Find(&thresholds).Error; err != nil {
		return fmt.Errorf("failed to get thresholds: %w", err)
	}

	for _, threshold := range thresholds {
		var currentValue float64

		switch threshold.Type {
		case metrics.CPUUsage:
			currentValue = currentMetrics.CPUUsage
		case metrics.MemoryUsage:
			currentValue = currentMetrics.MemoryUsage
		default:
			continue
		}

		// Check if threshold is breached
		if currentValue > threshold.Threshold {
			// Check if there's already an active alert for this type
			var existingAlert Alert
			err := s.db.Where("metric_type = ? AND status = ?", threshold.Type, AlertActive).
				First(&existingAlert).Error

			if err == gorm.ErrRecordNotFound {
				// Create new alert
				alert := Alert{
					Type:        threshold.Type,
					Message:     s.generateAlertMessage(threshold.Type, currentValue, threshold.Threshold),
					Value:       currentValue,
					Threshold:   threshold.Threshold,
					Severity:    s.calculateSeverity(currentValue, threshold.Threshold),
					Status:      AlertActive,
					TriggeredAt: currentMetrics.Timestamp,
				}

				if err := s.db.Create(&alert).Error; err != nil {
					log.Printf("Failed to create alert: %v", err)
				} else {
					log.Printf("Alert created: %s - %.2f%% > %.2f%%",
						threshold.Type, currentValue, threshold.Threshold)
				}
			}
		} else {
			// Resolve any active alerts for this type
			s.resolveActiveAlerts(threshold.Type)
		}
	}

	return nil
}

// resolveActiveAlerts resolves all active alerts for a specific metric type
func (s *Service) resolveActiveAlerts(metricType metrics.MetricType) {
	now := time.Now()
	result := s.db.Model(&Alert{}).
		Where("metric_type = ? AND status = ?", metricType, AlertActive).
		Updates(map[string]interface{}{
			"status":      AlertResolved,
			"resolved_at": &now,
		})

	if result.Error != nil {
		log.Printf("Failed to resolve alerts for %s: %v", metricType, result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("Resolved %d alerts for %s", result.RowsAffected, metricType)
	}
}

// generateAlertMessage creates a descriptive alert message
func (s *Service) generateAlertMessage(metricType metrics.MetricType, value, threshold float64) string {
	switch metricType {
	case metrics.CPUUsage:
		return fmt.Sprintf("High CPU usage detected: %.2f%% (threshold: %.2f%%)", value, threshold)
	case metrics.MemoryUsage:
		return fmt.Sprintf("High memory usage detected: %.2f%% (threshold: %.2f%%)", value, threshold)
	default:
		return fmt.Sprintf("Threshold breached for %s: %.2f%% (threshold: %.2f%%)", metricType, value, threshold)
	}
}

// calculateSeverity determines alert severity based on how much the threshold is exceeded
func (s *Service) calculateSeverity(value, threshold float64) AlertSeverity {
	exceedPercentage := ((value - threshold) / threshold) * 100

	switch {
	case exceedPercentage >= 50: // 50% above threshold
		return SeverityCritical
	case exceedPercentage >= 25: // 25% above threshold
		return SeverityHigh
	case exceedPercentage >= 10: // 10% above threshold
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// GetAlerts returns alerts with optional filtering
func (s *Service) GetAlerts(status AlertStatus, limit int) ([]Alert, error) {
	var alerts []Alert

	query := s.db.Order("triggered_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, nil
}

// GetAlertSummary returns comprehensive alert statistics
func (s *Service) GetAlertSummary(limit int) (*AlertSummary, error) {
	summary := &AlertSummary{
		AlertsByType:     make(map[metrics.MetricType]int64),
		AlertsBySeverity: make(map[AlertSeverity]int64),
	}

	// Get total alerts count
	if err := s.db.Model(&Alert{}).Count(&summary.TotalAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total alerts: %w", err)
	}

	// Get active alerts count
	if err := s.db.Model(&Alert{}).Where("status = ?", AlertActive).
		Count(&summary.ActiveAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active alerts: %w", err)
	}

	// Get resolved alerts count
	if err := s.db.Model(&Alert{}).Where("status = ?", AlertResolved).
		Count(&summary.ResolvedAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count resolved alerts: %w", err)
	}

	// Get alerts by type
	var typeResults []struct {
		Type  metrics.MetricType `json:"type"`
		Count int64              `json:"count"`
	}
	if err := s.db.Model(&Alert{}).
		Select("metric_type as type, COUNT(*) as count").
		Group("metric_type").
		Scan(&typeResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get alerts by type: %w", err)
	}

	for _, result := range typeResults {
		summary.AlertsByType[result.Type] = result.Count
	}

	// Get alerts by severity
	var severityResults []struct {
		Severity AlertSeverity `json:"severity"`
		Count    int64         `json:"count"`
	}
	if err := s.db.Model(&Alert{}).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Scan(&severityResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get alerts by severity: %w", err)
	}

	for _, result := range severityResults {
		summary.AlertsBySeverity[result.Severity] = result.Count
	}

	// Get recent alerts
	recentAlerts, err := s.GetAlerts("", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}
	summary.RecentAlerts = recentAlerts

	return summary, nil
}

// CreateAlert manually creates an alert (for testing purposes)
func (s *Service) CreateAlert(req *CreateAlertRequest) (*Alert, error) {
	alert := Alert{
		Type:        req.Type,
		Message:     s.generateAlertMessage(req.Type, req.Value, req.Threshold),
		Value:       req.Value,
		Threshold:   req.Threshold,
		Severity:    s.calculateSeverity(req.Value, req.Threshold),
		Status:      AlertActive,
		TriggeredAt: time.Now(),
	}

	if err := s.db.Create(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return &alert, nil
}

// ResolveAlert manually resolves an alert
func (s *Service) ResolveAlert(alertID uint) error {
	now := time.Now()
	result := s.db.Model(&Alert{}).
		Where("id = ? AND status = ?", alertID, AlertActive).
		Updates(map[string]interface{}{
			"status":      AlertResolved,
			"resolved_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve alert: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}
