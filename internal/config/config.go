package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	ServerAddress        string
	DatabaseURI          string
	AccrualSystemAddress string
}

// NewConfig creates a new configuration with values from flags and environment variables
func NewConfig() *Config {
	cfg := &Config{}

	// Define flags
	flag.StringVar(&cfg.ServerAddress, "a", "", "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")

	// Parse flags
	flag.Parse()

	// Override from environment variables if set
	if envVal := os.Getenv("RUN_ADDRESS"); envVal != "" {
		cfg.ServerAddress = envVal
	}

	if envVal := os.Getenv("DATABASE_URI"); envVal != "" {
		cfg.DatabaseURI = envVal
	}

	if envVal := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envVal != "" {
		cfg.AccrualSystemAddress = envVal
	}

	// Set defaults if not provided
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = "localhost:8080"
	}

	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"
	}

	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = "http://localhost:8081"
	}

	return cfg
}
