package repository

import (
	"context"
	"gophermart/domain/entity"
)

// WithdrawalRepository defines methods to work with withdrawals
type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *entity.Withdrawal) error
	GetByUserID(ctx context.Context, userID int64) ([]entity.Withdrawal, error)
}
