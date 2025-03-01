package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DBConfig contains database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB creates and initializes a new database connection
func NewDB(cfg DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize database schemas
	if err := initDB(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// initDB creates necessary tables if they don't exist
func initDB(ctx context.Context, db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(255) PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			status VARCHAR(50) NOT NULL,
			accrual DECIMAL(10, 2) DEFAULT 0,
			uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS balances (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			current DECIMAL(10, 2) NOT NULL DEFAULT 0,
			withdrawn DECIMAL(10, 2) NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			order_id VARCHAR(255) NOT NULL,
			sum DECIMAL(10, 2) NOT NULL,
			processed_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %s: %w", query, err)
		}
	}

	return nil
}
