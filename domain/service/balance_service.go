package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart/domain/entity"
	"gophermart/domain/repository"
)

// BalanceService handles balance-related business logic
type BalanceService struct {
	balanceRepo    repository.BalanceRepository
	withdrawalRepo repository.WithdrawalRepository
	orderRepo      repository.OrderRepository
}

// NewBalanceService creates a new BalanceService
func NewBalanceService(
	balanceRepo repository.BalanceRepository,
	withdrawalRepo repository.WithdrawalRepository,
	orderRepo repository.OrderRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		orderRepo:      orderRepo,
	}
}

// GetUserBalance retrieves a user's balance
func (s *BalanceService) GetUserBalance(ctx context.Context, userID int64) (*entity.Balance, error) {
	return s.balanceRepo.GetOrCreate(ctx, userID)
}

// WithdrawPoints withdraws points from a user's balance
func (s *BalanceService) WithdrawPoints(ctx context.Context, userID int64, orderID string, amount float64) error {
	// Validate order number
	if !ValidateLuhn(orderID) {
		return errors.New("invalid order number")
	}

	// Check if order already exists
	exists, _, err := s.orderRepo.CheckExists(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to check order existence: %w", err)
	}

	if exists {
		return errors.New("order already exists")
	}

	// Update balance
	if err := s.balanceRepo.UpdateBalance(ctx, userID, amount, true); err != nil {
		return fmt.Errorf("failed to withdraw points: %w", err)
	}

	// Create withdrawal record
	withdrawal := &entity.Withdrawal{
		UserID:  userID,
		OrderID: orderID,
		Sum:     amount,
	}

	if err := s.withdrawalRepo.Create(ctx, withdrawal); err != nil {
		return fmt.Errorf("failed to create withdrawal record: %w", err)
	}

	return nil
}

// GetUserWithdrawals retrieves all withdrawals for a user
func (s *BalanceService) GetUserWithdrawals(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	return s.withdrawalRepo.GetByUserID(ctx, userID)
}
