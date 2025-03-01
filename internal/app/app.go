package app

import (
	"context"
	"fmt"
	"gophermart/domain/service"
	"gophermart/internal/http"
	"time"
)

// App represents the main application
type App struct {
	server         *http.Server
	accrualService *service.AccrualService
}

// NewApp creates a new application
func NewApp(server *http.Server, accrualService *service.AccrualService) *App {
	return &App{
		server:         server,
		accrualService: accrualService,
	}
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	// Start accrual service
	a.accrualService.Start(ctx)

	// Run HTTP server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		fmt.Println("Starting server...")
		errCh <- a.server.Start()
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		fmt.Println("Shutting down gracefully...")

		// Create a timeout context for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Stop accrual service
		a.accrualService.Stop()

		// Shutdown the server
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error during server shutdown: %w", err)
		}

		return nil
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
}
