package repository

import (
	"context"
	"gophermart/domain/entity"
)

// BalanceRepository defines methods to work with balance
type BalanceRepository interface {
	GetOrCreate(ctx context.Context, userID int64) (*entity.Balance, error)
	UpdateBalance(ctx context.Context, userID int64, amount float64, isWithdrawal bool) error
}
