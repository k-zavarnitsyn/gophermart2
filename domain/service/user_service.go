package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart/domain/entity"
	"gophermart/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Register registers a new user
func (s *UserService) Register(ctx context.Context, login, password string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByLogin(ctx, login)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	user := &entity.User{
		Login:    login,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *UserService) Login(ctx context.Context, login, password string) (*entity.User, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
