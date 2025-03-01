package entity

import "time"

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"-"` // Password hash, not exposed in JSON
	CreatedAt time.Time `json:"created_at"`
}

// Order represents an order in the system
type Order struct {
	ID         string    `json:"id"`
	UserID     int64     `json:"user_id"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance represents user's loyalty balance
type Balance struct {
	UserID    int64     `json:"user_id"`
	Current   float64   `json:"current"`
	Withdrawn float64   `json:"withdrawn"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Withdrawal represents a withdrawal operation
type Withdrawal struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	OrderID     string    `json:"order_id"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// OrderStatus represents possible order statuses
const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)
