package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/domain/entity"
)

// OrderRepo implements the OrderRepository interface
type OrderRepo struct {
	db *sql.DB
}

// NewOrderRepo creates a new OrderRepo instance
func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create adds a new order to the database
func (r *OrderRepo) Create(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (id, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING uploaded_at
	`

	err := r.db.QueryRowContext(ctx, query, order.ID, order.UserID, order.Status).Scan(&order.UploadedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *OrderRepo) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	query := `
		SELECT id, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE id = $1
	`

	order := &entity.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	return order, nil
}

// GetByUserID retrieves all orders for a user
func (r *OrderRepo) GetByUserID(ctx context.Context, userID int64) ([]entity.Order, error) {
	query := `
		SELECT id, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	return orders, nil
}

// Update updates an existing order
func (r *OrderRepo) Update(ctx context.Context, order *entity.Order) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, order.Status, order.Accrual, order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// CheckExists checks if an order exists and returns the user ID if it does
func (r *OrderRepo) CheckExists(ctx context.Context, id string) (bool, int64, error) {
	query := `
		SELECT user_id FROM orders WHERE id = $1
	`

	var userID int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("failed to check if order exists: %w", err)
	}

	return true, userID, nil
}
