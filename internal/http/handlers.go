package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// UserCredentials represents user login/register credentials
type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// WithdrawalRequest represents a withdrawal request
type WithdrawalRequest struct {
	OrderID string  `json:"order"`
	Sum     float64 `json:"sum"`
}

// register handles user registration
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := s.userService.Register(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 1 day
	})

	w.WriteHeader(http.StatusOK)
}

// login handles user login
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := s.userService.Login(r.Context(), creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set token in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 1 day
	})

	w.WriteHeader(http.StatusOK)
}

// handleOrders handles getting and uploading orders
func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request, userID int64) {
	switch r.Method {
	case http.MethodGet:
		s.getOrders(w, r, userID)
	case http.MethodPost:
		s.uploadOrder(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getOrders retrieves all orders for a user
func (s *Server) getOrders(w http.ResponseWriter, r *http.Request, userID int64) {
	orders, err := s.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// uploadOrder uploads a new order
func (s *Server) uploadOrder(w http.ResponseWriter, r *http.Request, userID int64) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	orderID := strings.TrimSpace(string(body))
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	_, err = s.orderService.UploadOrder(r.Context(), orderID, userID)
	if err != nil {
		switch {
		case err.Error() == "invalid order number":
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		case err.Error() == "order already uploaded by another user":
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		case err.Error() == "order already uploaded by you":
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// getBalance retrieves a user's balance
func (s *Server) getBalance(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balance, err := s.balanceService.GetUserBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	response := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// withdraw handles point withdrawal
func (s *Server) withdraw(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" || req.Sum <= 0 {
		http.Error(w, "Invalid withdrawal request", http.StatusBadRequest)
		return
	}

	err := s.balanceService.WithdrawPoints(r.Context(), userID, req.OrderID, req.Sum)
	if err != nil {
		switch {
		case err.Error() == "invalid order number":
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		case err.Error() == "order already exists":
			http.Error(w, "Order already exists", http.StatusConflict)
		case err.Error() == "insufficient funds":
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// getWithdrawals retrieves all withdrawals for a user
func (s *Server) getWithdrawals(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	withdrawals, err := s.balanceService.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get withdrawals", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// withAuth is a middleware to authenticate requests
func (s *Server) withAuth(handler func(http.ResponseWriter, *http.Request, int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		userID, err := validateToken(cookie.Value)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handler(w, r, userID)
	}
}
