package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	SessionDuration time.Duration `mapstructure:"session_duration"`
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
	CPUThreshold       float64       `mapstructure:"cpu_threshold"`
	MemoryThreshold    float64       `mapstructure:"memory_threshold"`
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Set default values first
	setDefaults()

	// Set up Viper to read .env file
	viper.SetConfigName(".env")
	viper.SetConfigType("dotenv")
	viper.AddConfigPath(".")

	// Read .env file
	if err := viper.ReadInConfig(); err != nil {
		// .env file is optional, continue if not found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading .env file: %w", err)
		}
	}

	// Enable automatic environment variable reading
	viper.AutomaticEnv()

	// Map environment variables to config structure
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("PORT")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("ACCESS_TOKEN_SECRET")
	viper.BindEnv("CPU_THRESHOLD")
	viper.BindEnv("MEMORY_THRESHOLD")

	// Create config with direct viper calls
	config := &Config{
		Server: ServerConfig{
			Port:         viper.GetString("PORT"),
			Host:         viper.GetString("HOST"),
			ReadTimeout:  viper.GetDuration("server.read_timeout"),
			WriteTimeout: viper.GetDuration("server.write_timeout"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		Auth: AuthConfig{
			JWTSecret:       getJWTSecret(),
			SessionDuration: viper.GetDuration("auth.session_duration"),
		},
		Metrics: MetricsConfig{
			CollectionInterval: viper.GetDuration("metrics.collection_interval"),
			CPUThreshold:       viper.GetFloat64("CPU_THRESHOLD"),
			MemoryThreshold:    viper.GetFloat64("MEMORY_THRESHOLD"),
		},
	}

	// Apply defaults if values are empty
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Auth.JWTSecret == "" {
		config.Auth.JWTSecret = "your-secret-key"
	}
	if config.Metrics.CPUThreshold == 0 {
		config.Metrics.CPUThreshold = 80.0
	}
	if config.Metrics.MemoryThreshold == 0 {
		config.Metrics.MemoryThreshold = 75.0
	}

	return config, nil
}

// getJWTSecret tries multiple environment variables for JWT secret
func getJWTSecret() string {
	if secret := viper.GetString("JWT_SECRET"); secret != "" {
		return secret
	}
	if secret := viper.GetString("ACCESS_TOKEN_SECRET"); secret != "" {
		return secret
	}
	return ""
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")

	// Database defaults
	viper.SetDefault("database.url", "")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-secret-key")
	viper.SetDefault("auth.session_duration", "24h")

	// Metrics defaults
	viper.SetDefault("metrics.collection_interval", "30s")
	viper.SetDefault("metrics.cpu_threshold", 80.0)
	viper.SetDefault("metrics.memory_threshold", 75.0)
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return c.Database.URL
}
