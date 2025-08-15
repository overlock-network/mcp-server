package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	// API Configuration
	OverlockGRPCURL string // gRPC endpoint URL
	APITimeout      time.Duration

	// Server Configuration
	HTTPAddr string

	// Debug Configuration
	Debug bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		// Default values
		OverlockGRPCURL: "localhost:9090", // gRPC endpoint
		APITimeout:      30 * time.Second,
		HTTPAddr:        "127.0.0.1:8080",
		Debug:           false,
	}

	// Load from environment variables
	if grpcURL := os.Getenv("OVERLOCK_GRPC_URL"); grpcURL != "" {
		config.OverlockGRPCURL = grpcURL
	}

	if timeout := os.Getenv("OVERLOCK_API_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.APITimeout = d
		} else {
			log.Printf("Warning: Invalid OVERLOCK_API_TIMEOUT '%s', using default %v: %v", timeout, config.APITimeout, err)
		}
	}

	if addr := os.Getenv("MCP_HTTP_ADDR"); addr != "" {
		config.HTTPAddr = addr
	}

	if debug := os.Getenv("DEBUG"); debug == "true" {
		config.Debug = true
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.OverlockGRPCURL == "" {
		return fmt.Errorf("OVERLOCK_GRPC_URL is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("MCP_HTTP_ADDR is required")
	}
	if c.APITimeout <= 0 {
		return fmt.Errorf("OVERLOCK_API_TIMEOUT must be positive")
	}
	return nil
}
