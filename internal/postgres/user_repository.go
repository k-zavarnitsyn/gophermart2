package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart/domain/entity"
)

// UserRepo implements the UserRepository interface
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo instance
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create adds a new user to the database
func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (login, password)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, user.Login, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByLogin retrieves a user by login
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	query := `
		SELECT id, login, password, created_at
		FROM users
		WHERE login = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT id, login, password, created_at
		FROM users
		WHERE id = $1
	`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}
