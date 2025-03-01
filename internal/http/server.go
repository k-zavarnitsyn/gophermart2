package http

import (
	"context"
	"gophermart/domain/service"
	"net/http"
	"time"
)

// Server represents the HTTP server
type Server struct {
	server         *http.Server
	userService    *service.UserService
	orderService   *service.OrderService
	balanceService *service.BalanceService
}

// NewServer creates a new HTTP server
func NewServer(
	addr string,
	userService *service.UserService,
	orderService *service.OrderService,
	balanceService *service.BalanceService,
) *Server {
	server := &Server{
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
	}

	mux := http.NewServeMux()

	// User endpoints
	mux.HandleFunc("/api/user/register", server.register)
	mux.HandleFunc("/api/user/login", server.login)

	// Order endpoints
	mux.HandleFunc("/api/user/orders", server.withAuth(server.handleOrders))

	// Balance endpoints
	mux.HandleFunc("/api/user/balance", server.withAuth(server.getBalance))
	mux.HandleFunc("/api/user/balance/withdraw", server.withAuth(server.withdraw))
	mux.HandleFunc("/api/user/withdrawals", server.withAuth(server.getWithdrawals))

	server.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
