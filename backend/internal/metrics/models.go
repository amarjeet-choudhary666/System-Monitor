package metrics

import (
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	CPUUsage    MetricType = "cpu_usage"
	MemoryUsage MetricType = "memory_usage"
)

// Metric represents a system metric reading
type Metric struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Type      MetricType `json:"type" gorm:"column:metric_type"`
	Value     float64    `json:"value" gorm:"not null"`
	Unit      string     `json:"unit" gorm:"not null"`
	Timestamp time.Time  `json:"timestamp" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at"`
}

// SystemMetrics represents current system metrics
type SystemMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	Timestamp   time.Time `json:"timestamp"`
}

type MetricThreshold struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Type      MetricType `json:"type" gorm:"column:metric_type;unique"`
	Threshold float64    `json:"threshold" gorm:"not null"`
	Enabled   bool       `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// MetricSummary represents aggregated metric data
type MetricSummary struct {
	Type    MetricType `json:"type"`
	Average float64    `json:"average"`
	Min     float64    `json:"min"`
	Max     float64    `json:"max"`
	Count   int64      `json:"count"`
}
