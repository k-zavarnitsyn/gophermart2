package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gophermart/domain/entity"
	"gophermart/domain/repository"
	"net/http"
	"sync"
	"time"
)

// AccrualResponse represents the response from the accrual system
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// AccrualService handles interaction with the accrual system
type AccrualService struct {
	orderRepo    repository.OrderRepository
	accrualURL   string
	client       *http.Client
	pollInterval time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewAccrualService creates a new AccrualService
func NewAccrualService(orderRepo repository.OrderRepository, accrualURL string, pollInterval time.Duration) *AccrualService {
	return &AccrualService{
		orderRepo:  orderRepo,
		accrualURL: accrualURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

// Start starts the accrual service
func (s *AccrualService) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.pollAccrualSystem(ctx)
}

// Stop stops the accrual service
func (s *AccrualService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// pollAccrualSystem periodically checks for new orders and updates their status
func (s *AccrualService) pollAccrualSystem(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processNewOrders(ctx)
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// processNewOrders processes all new orders
func (s *AccrualService) processNewOrders(ctx context.Context) {
	// Get all orders with status NEW or PROCESSING
	orders, err := s.getOrdersToProcess(ctx)
	if err != nil {
		fmt.Printf("Failed to get orders to process: %v\n", err)
		return
	}

	// Process each order
	for _, order := range orders {
		status, accrual, err := s.checkOrderStatus(ctx, order.ID)
		if err != nil {
			fmt.Printf("Failed to check order status for order %s: %v\n", order.ID, err)
			continue
		}

		// Update order status if changed
		if status != order.Status {
			err := s.updateOrderStatus(ctx, order.ID, status, accrual)
			if err != nil {
				fmt.Printf("Failed to update order status for order %s: %v\n", order.ID, err)
			}
		}
	}
}

// getOrdersToProcess retrieves all orders that need processing
func (s *AccrualService) getOrdersToProcess(ctx context.Context) ([]entity.Order, error) {
	// This is a simplified implementation. In a real system, you would need to:
	// 1. Query the database for orders with status NEW or PROCESSING
	// 2. Implement pagination or batching for large datasets
	// 3. Handle rate limiting for the accrual system

	// For this example, just returning an empty slice
	return []entity.Order{}, nil
}

// checkOrderStatus checks the status of an order in the accrual system
func (s *AccrualService) checkOrderStatus(ctx context.Context, orderID string) (string, float64, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle different response codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Process successful response
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return "", 0, fmt.Errorf("failed to decode response: %w", err)
		}

		return accrualResp.Status, accrualResp.Accrual, nil

	case http.StatusTooManyRequests:
		// Handle rate limiting
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			// Parse the retry-after header and wait
			seconds, err := time.ParseDuration(retryAfter + "s")
			if err == nil {
				time.Sleep(seconds)
			}
		}

		return "", 0, fmt.Errorf("rate limited by accrual system")

	case http.StatusNoContent:
		// Order not found in accrual system
		return entity.StatusInvalid, 0, nil

	default:
		return "", 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// updateOrderStatus updates the status of an order
func (s *AccrualService) updateOrderStatus(ctx context.Context, orderID, status string, accrual float64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = status
	order.Accrual = accrual

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// CheckOrderDirectly checks the status of an order directly (can be called from API)
func (s *AccrualService) CheckOrderDirectly(ctx context.Context, orderID string) (string, float64, error) {
	return s.checkOrderStatus(ctx, orderID)
}
