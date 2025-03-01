// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"gophermart/domain/service"
	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/internal/http"
	"gophermart/internal/postgres"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Parse database URI
	dbURL, err := url.Parse(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Invalid database URI: %v", err)
	}

	// Extract database parameters
	dbConfig := postgres.DBConfig{
		Host: dbURL.Hostname(),
		Port: dbURL.Port(),
		User: dbURL.User.Username(),
		Password: func() string {
			pass, _ := dbURL.User.Password()
			return pass
		}(),
		DBName: dbURL.Path[1:],
		SSLMode: func() string {
			q := dbURL.Query()
			return q.Get("sslmode")
		}(),
	}

	// Initialize database
	db, err := postgres.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create repositories
	userRepo := postgres.NewUserRepo(db)
	orderRepo := postgres.NewOrderRepo(db)
	balanceRepo := postgres.NewBalanceRepo(db)
	withdrawalRepo := postgres.NewWithdrawalRepo(db)

	// Create services
	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo, balanceRepo)
	balanceService := service.NewBalanceService(balanceRepo, withdrawalRepo, orderRepo)
	accrualService := service.NewAccrualService(orderRepo, cfg.AccrualSystemAddress, 1*time.Minute)

	// Create HTTP server
	server := http.NewServer(cfg.ServerAddress, userService, orderService, balanceService)

	// Create application
	app := app.NewApp(server, accrualService)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Listen for interrupt signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	// Start the application
	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
