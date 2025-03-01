package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"gophermart/domain/entity"
)

// WithdrawalRepo implements the WithdrawalRepository interface
type WithdrawalRepo struct {
	db *sql.DB
}

// NewWithdrawalRepo creates a new WithdrawalRepo instance
func NewWithdrawalRepo(db *sql.DB) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

// Create adds a new withdrawal record
func (r *WithdrawalRepo) Create(ctx context.Context, withdrawal *entity.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, order_id, sum)
		VALUES ($1, $2, $3)
		RETURNING id, processed_at
	`

	err := r.db.QueryRowContext(ctx, query, withdrawal.UserID, withdrawal.OrderID, withdrawal.Sum).Scan(
		&withdrawal.ID,
		&withdrawal.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

// GetByUserID retrieves all withdrawals for a user
func (r *WithdrawalRepo) GetByUserID(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	query := `
		SELECT id, user_id, order_id, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []entity.Withdrawal
	for rows.Next() {
		var w entity.Withdrawal
		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderID,
			&w.Sum,
			&w.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal row: %w", err)
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating withdrawal rows: %w", err)
	}

	return withdrawals, nil
}
