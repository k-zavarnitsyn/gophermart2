package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart/domain/entity"
	"gophermart/domain/repository"
	"strconv"
)

// OrderService handles order-related business logic
type OrderService struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
}

// NewOrderService creates a new OrderService
func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
	}
}

// UploadOrder uploads a new order
func (s *OrderService) UploadOrder(ctx context.Context, orderID string, userID int64) (*entity.Order, error) {
	// Validate order number using Luhn algorithm
	if !ValidateLuhn(orderID) {
		return nil, errors.New("invalid order number")
	}

	// Check if order already exists
	exists, existingUserID, err := s.orderRepo.CheckExists(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	// If order exists and belongs to another user
	if exists && existingUserID != userID {
		return nil, errors.New("order already uploaded by another user")
	}

	// If order exists and belongs to the current user
	if exists && existingUserID == userID {
		return nil, errors.New("order already uploaded by you")
	}

	// Create new order
	order := &entity.Order{
		ID:     orderID,
		UserID: userID,
		Status: entity.StatusNew,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]entity.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID)
}

// UpdateOrderStatus updates the status and accrual of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, status string, accrual float64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Only update if status changed
	if order.Status != status {
		order.Status = status
		order.Accrual = accrual

		// If order processed successfully, update user balance
		if status == entity.StatusProcessed && accrual > 0 {
			err := s.balanceRepo.UpdateBalance(ctx, order.UserID, accrual, false)
			if err != nil {
				return fmt.Errorf("failed to update balance: %w", err)
			}
		}

		// Update order in the database
		if err := s.orderRepo.Update(ctx, order); err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
	}

	return nil
}

// ValidateLuhn validates a number using the Luhn algorithm
func ValidateLuhn(number string) bool {
	digits := make([]int, len(number))
	for i, r := range number {
		digit, err := strconv.Atoi(string(r))
		if err != nil {
			return false
		}
		digits[i] = digit
	}

	checksum := 0
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if (len(digits)-i)%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		checksum += d
	}

	return checksum%10 == 0
}
