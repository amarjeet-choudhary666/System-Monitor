package alerts

import (
	"time"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/metrics"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertActive   AlertStatus = "active"
	AlertResolved AlertStatus = "resolved"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"
	SeverityMedium   AlertSeverity = "medium"
	SeverityHigh     AlertSeverity = "high"
	SeverityCritical AlertSeverity = "critical"
)

// Alert represents a system alert
type Alert struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	Type        metrics.MetricType `json:"type" gorm:"column:metric_type"`
	Message     string             `json:"message" gorm:"not null"`
	Value       float64            `json:"value" gorm:"not null"`
	Threshold   float64            `json:"threshold" gorm:"not null"`
	Severity    AlertSeverity      `json:"severity" gorm:"not null"`
	Status      AlertStatus        `json:"status" gorm:"default:'active'"`
	TriggeredAt time.Time          `json:"triggered_at" gorm:"not null"`
	ResolvedAt  *time.Time         `json:"resolved_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// AlertSummary represents aggregated alert statistics
type AlertSummary struct {
	TotalAlerts      int64                        `json:"total_alerts"`
	ActiveAlerts     int64                        `json:"active_alerts"`
	ResolvedAlerts   int64                        `json:"resolved_alerts"`
	AlertsByType     map[metrics.MetricType]int64 `json:"alerts_by_type"`
	AlertsBySeverity map[AlertSeverity]int64      `json:"alerts_by_severity"`
	RecentAlerts     []Alert                      `json:"recent_alerts"`
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Type      metrics.MetricType `json:"type" binding:"required"`
	Value     float64            `json:"value" binding:"required"`
	Threshold float64            `json:"threshold" binding:"required"`
}
