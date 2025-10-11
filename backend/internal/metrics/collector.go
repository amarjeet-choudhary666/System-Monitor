package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
)

// Collector handles system metrics collection
type Collector struct {
	db       *gorm.DB
	interval time.Duration
	stopCh   chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(db *gorm.DB, interval time.Duration) *Collector {
	return &Collector{
		db:       db,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins collecting metrics at regular intervals
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	log.Printf("Starting metrics collection with interval: %v", c.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Metrics collection stopped by context")
			return
		case <-c.stopCh:
			log.Println("Metrics collection stopped")
			return
		case <-ticker.C:
			if err := c.collectMetrics(); err != nil {
				log.Printf("Error collecting metrics: %v", err)
			}
		}
	}
}

// Stop stops the metrics collection
func (c *Collector) Stop() {
	close(c.stopCh)
}

// collectMetrics collects current system metrics
func (c *Collector) collectMetrics() error {
	now := time.Now()

	// Collect CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("failed to get CPU usage: %w", err)
	}

	if len(cpuPercent) > 0 {
		cpuMetric := Metric{
			Type:      CPUUsage,
			Value:     cpuPercent[0],
			Unit:      "%",
			Timestamp: now,
		}

		if err := c.db.Create(&cpuMetric).Error; err != nil {
			log.Printf("Failed to save CPU metric: %v", err)
		}
	}

	// Collect Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory usage: %w", err)
	}

	memoryMetric := Metric{
		Type:      MemoryUsage,
		Value:     memInfo.UsedPercent,
		Unit:      "%",
		Timestamp: now,
	}

	if err := c.db.Create(&memoryMetric).Error; err != nil {
		log.Printf("Failed to save memory metric: %v", err)
	}

	log.Printf("Collected metrics - CPU: %.2f%%, Memory: %.2f%%",
		cpuPercent[0], memInfo.UsedPercent)

	return nil
}

// GetCurrentMetrics returns the latest system metrics
func (c *Collector) GetCurrentMetrics() (*SystemMetrics, error) {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}

	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	return &SystemMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memInfo.UsedPercent,
		Timestamp:   time.Now(),
	}, nil
}

// GetMetricHistory returns historical metrics for a specific type
func (c *Collector) GetMetricHistory(metricType MetricType, limit int) ([]Metric, error) {
	var metrics []Metric

	query := c.db.Where("metric_type = ?", metricType).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("failed to get metric history: %w", err)
	}

	return metrics, nil
}

// GetMetricSummary returns aggregated metrics for the last N readings
func (c *Collector) GetMetricSummary(metricType MetricType, limit int) (*MetricSummary, error) {
	var result struct {
		Average float64
		Min     float64
		Max     float64
		Count   int64
	}

	query := c.db.Model(&Metric{}).
		Select("AVG(value) as average, MIN(value) as min, MAX(value) as max, COUNT(*) as count").
		Where("metric_type = ?", metricType)

	if limit > 0 {
		// Get the last N records by timestamp
		subQuery := c.db.Model(&Metric{}).
			Select("id").
			Where("metric_type = ?", metricType).
			Order("timestamp DESC").
			Limit(limit)

		query = query.Where("id IN (?)", subQuery)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to get metric summary: %w", err)
	}

	return &MetricSummary{
		Type:    metricType,
		Average: result.Average,
		Min:     result.Min,
		Max:     result.Max,
		Count:   result.Count,
	}, nil
}

// InitializeThresholds sets up default metric thresholds
func (c *Collector) InitializeThresholds() error {
	thresholds := []MetricThreshold{
		{Type: CPUUsage, Threshold: 80.0, Enabled: true},
		{Type: MemoryUsage, Threshold: 75.0, Enabled: true},
	}

	for _, threshold := range thresholds {
		// Use raw SQL to avoid cached plan issues
		var count int64
		err := c.db.Raw("SELECT COUNT(*) FROM metric_thresholds WHERE metric_type = ?", threshold.Type).Scan(&count).Error
		if err != nil {
			return fmt.Errorf("failed to check existing threshold: %w", err)
		}

		if count == 0 {
			// Create new threshold using raw SQL
			err := c.db.Exec(`
				INSERT INTO metric_thresholds (metric_type, threshold, enabled, created_at, updated_at) 
				VALUES (?, ?, ?, NOW(), NOW())
			`, threshold.Type, threshold.Threshold, threshold.Enabled).Error

			if err != nil {
				return fmt.Errorf("failed to create threshold for %s: %w", threshold.Type, err)
			}
			log.Printf("Created default threshold for %s: %.1f%%", threshold.Type, threshold.Threshold)
		}
	}

	return nil
}
