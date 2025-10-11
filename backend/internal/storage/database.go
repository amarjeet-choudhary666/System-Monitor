package storage

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/alerts"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/auth"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/config"
	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/metrics"
)

// Database holds the database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	var db *gorm.DB
	var err error

	// Check if it's an in-memory SQLite database (for testing)
	if dsn == ":memory:" {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
		}
		log.Println("Successfully connected to in-memory SQLite database")
	} else {
		// Use PostgreSQL driver for DATABASE_URL
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
		}

		// Test the connection
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %w", err)
		}

		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		log.Println("Successfully connected to PostgreSQL database")
	}

	return &Database{DB: db}, nil
}

// AutoMigrate runs database migrations
func (d *Database) AutoMigrate() error {
	log.Println("Running database migrations...")

	// First, run the basic migrations
	err := d.DB.AutoMigrate(
		&auth.User{},
		&auth.Session{},
		&metrics.Metric{},
		&metrics.MetricThreshold{},
		&alerts.Alert{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Fix any existing NULL values in metric_type columns
	if err := d.fixMetricTypeColumns(); err != nil {
		log.Printf("Warning: Failed to fix metric_type columns: %v", err)
	}

	// Clear any cached query plans by closing and reopening the connection
	if err := d.refreshConnection(); err != nil {
		log.Printf("Warning: Failed to refresh database connection: %v", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// fixMetricTypeColumns updates any NULL values in metric_type columns and drops old type columns
func (d *Database) fixMetricTypeColumns() error {
	// Fix metric_thresholds table
	result := d.DB.Exec(`
		UPDATE metric_thresholds 
		SET metric_type = CASE 
			WHEN threshold = 80.0 THEN 'cpu_usage'
			WHEN threshold = 75.0 THEN 'memory_usage'
			ELSE 'cpu_usage'
		END 
		WHERE metric_type IS NULL OR metric_type = ''
	`)
	if result.Error != nil {
		log.Printf("Failed to fix metric_thresholds: %v", result.Error)
	}

	// Fix metrics table - set a default type for any NULL values
	result = d.DB.Exec(`
		UPDATE metrics 
		SET metric_type = 'cpu_usage' 
		WHERE metric_type IS NULL OR metric_type = ''
	`)
	if result.Error != nil {
		log.Printf("Failed to fix metrics: %v", result.Error)
	}

	// Fix alerts table
	result = d.DB.Exec(`
		UPDATE alerts 
		SET metric_type = CASE 
			WHEN message LIKE '%CPU%' OR message LIKE '%cpu%' THEN 'cpu_usage'
			WHEN message LIKE '%memory%' OR message LIKE '%Memory%' THEN 'memory_usage'
			ELSE 'cpu_usage'
		END 
		WHERE metric_type IS NULL OR metric_type = ''
	`)
	if result.Error != nil {
		log.Printf("Failed to fix alerts: %v", result.Error)
	}

	// Drop old type columns if they exist
	d.dropOldTypeColumns()

	return nil
}

// dropOldTypeColumns removes the old type columns that conflict with metric_type
func (d *Database) dropOldTypeColumns() {
	// Drop problematic columns from metrics table
	metricsColumns := []string{"type", "cpu_usage", "memory_usage"}
	for _, column := range metricsColumns {
		d.dropColumnIfExists("metrics", column)
	}

	// Drop type columns from other tables
	d.dropColumnIfExists("alerts", "type")
	d.dropColumnIfExists("metric_thresholds", "type")
}

// dropColumnIfExists drops a column if it exists
func (d *Database) dropColumnIfExists(table, column string) {
	var count int64
	result := d.DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = ? AND column_name = ? AND table_schema = CURRENT_SCHEMA()
	`, table, column).Scan(&count)

	if result.Error != nil {
		log.Printf("Failed to check for %s column in %s: %v", column, table, result.Error)
		return
	}

	if count > 0 {
		dropSQL := fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s", table, column)
		result = d.DB.Exec(dropSQL)
		if result.Error != nil {
			log.Printf("Failed to drop %s column from %s: %v", column, table, result.Error)
		} else {
			log.Printf("Dropped old %s column from %s table", column, table)
		}
	}
}

// refreshConnection closes and reopens the database connection to clear cached plans
func (d *Database) refreshConnection() error {
	// Get the underlying sql.DB
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Close all connections in the pool
	sqlDB.SetMaxOpenConns(0)
	sqlDB.SetMaxOpenConns(10) // Reset to a reasonable default

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}
