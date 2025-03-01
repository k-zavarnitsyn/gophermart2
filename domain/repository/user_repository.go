package repository

import (
	"context"
	"gophermart/domain/entity"
)

// UserRepository defines methods to work with users
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByLogin(ctx context.Context, login string) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
}
