package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/domain/entity"
	"time"
)

// BalanceRepo implements the BalanceRepository interface
type BalanceRepo struct {
	db *sql.DB
}

// NewBalanceRepo creates a new BalanceRepo instance
func NewBalanceRepo(db *sql.DB) *BalanceRepo {
	return &BalanceRepo{db: db}
}

// GetOrCreate retrieves or creates a balance record for a user
func (r *BalanceRepo) GetOrCreate(ctx context.Context, userID int64) (*entity.Balance, error) {
	// Try to get existing balance
	query := `
		SELECT user_id, current, withdrawn, updated_at
		FROM balances
		WHERE user_id = $1
	`

	balance := &entity.Balance{UserID: userID}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Current,
		&balance.Withdrawn,
		&balance.UpdatedAt,
	)

	if err == nil {
		return balance, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Create new balance if not exists
	insertQuery := `
		INSERT INTO balances (user_id, current, withdrawn)
		VALUES ($1, 0, 0)
		RETURNING current, withdrawn, updated_at
	`

	err = r.db.QueryRowContext(ctx, insertQuery, userID).Scan(
		&balance.Current,
		&balance.Withdrawn,
		&balance.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance: %w", err)
	}

	return balance, nil
}

// UpdateBalance updates a user's balance
func (r *BalanceRepo) UpdateBalance(ctx context.Context, userID int64, amount float64, isWithdrawal bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock the row for update
	query := `
		SELECT current, withdrawn FROM balances
		WHERE user_id = $1
		FOR UPDATE
	`

	var current, withdrawn float64
	err = tx.QueryRowContext(ctx, query, userID).Scan(&current, &withdrawn)
	if err != nil {
		return fmt.Errorf("failed to lock balance row: %w", err)
	}

	// Check sufficient funds for withdrawal
	if isWithdrawal && current < amount {
		return errors.New("insufficient funds")
	}

	// Update based on operation type
	var updateQuery string
	if isWithdrawal {
		updateQuery = `
			UPDATE balances
			SET current = current - $1, withdrawn = withdrawn + $1, updated_at = $2
			WHERE user_id = $3
		`
	} else {
		updateQuery = `
			UPDATE balances
			SET current = current + $1, updated_at = $2
			WHERE user_id = $3
		`
	}

	now := time.Now()
	_, err = tx.ExecContext(ctx, updateQuery, amount, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return tx.Commit()
}
