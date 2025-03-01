package repository

import (
	"context"
	"gophermart/domain/entity"
)

// OrderRepository defines methods to work with orders
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	GetByUserID(ctx context.Context, userID int64) ([]entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	CheckExists(ctx context.Context, id string) (bool, int64, error)
}
