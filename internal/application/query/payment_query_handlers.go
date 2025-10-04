package query

import (
	"context"
	"fmt"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
)

// GetPaymentQuery represents a query to get a payment by ID
type GetPaymentQuery struct {
	PaymentID string `json:"payment_id"`
}

// GetPaymentHandler handles get payment queries
type GetPaymentHandler struct {
	paymentRepo repository.PaymentRepository
}

// NewGetPaymentHandler creates a new get payment handler
func NewGetPaymentHandler(paymentRepo repository.PaymentRepository) *GetPaymentHandler {
	return &GetPaymentHandler{
		paymentRepo: paymentRepo,
	}
}

// Handle processes the get payment query
func (h *GetPaymentHandler) Handle(ctx context.Context, query *GetPaymentQuery) (*aggregate.Payment, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	if query.PaymentID == "" {
		return nil, fmt.Errorf("payment_id is required")
	}

	return h.paymentRepo.GetByID(ctx, query.PaymentID)
}

// GetPaymentByOrderCodeQuery represents a query to get a payment by order code
type GetPaymentByOrderCodeQuery struct {
	OrderCode int64 `json:"order_code"`
}

// GetPaymentByOrderCodeHandler handles get payment by order code queries
type GetPaymentByOrderCodeHandler struct {
	paymentRepo repository.PaymentRepository
}

// NewGetPaymentByOrderCodeHandler creates a new get payment by order code handler
func NewGetPaymentByOrderCodeHandler(paymentRepo repository.PaymentRepository) *GetPaymentByOrderCodeHandler {
	return &GetPaymentByOrderCodeHandler{
		paymentRepo: paymentRepo,
	}
}

// Handle processes the get payment by order code query
func (h *GetPaymentByOrderCodeHandler) Handle(ctx context.Context, query *GetPaymentByOrderCodeQuery) (*aggregate.Payment, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	if query.OrderCode == 0 {
		return nil, fmt.Errorf("order_code is required")
	}

	return h.paymentRepo.GetByOrderCode(ctx, query.OrderCode)
}

// ListUserPaymentsQuery represents a query to list payments for a user
type ListUserPaymentsQuery struct {
	UserID string `json:"user_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// ListUserPaymentsHandler handles list user payments queries
type ListUserPaymentsHandler struct {
	paymentRepo repository.PaymentRepository
}

// NewListUserPaymentsHandler creates a new list user payments handler
func NewListUserPaymentsHandler(paymentRepo repository.PaymentRepository) *ListUserPaymentsHandler {
	return &ListUserPaymentsHandler{
		paymentRepo: paymentRepo,
	}
}

// Handle processes the list user payments query
func (h *ListUserPaymentsHandler) Handle(ctx context.Context, query *ListUserPaymentsQuery) ([]*aggregate.Payment, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	if query.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if query.Limit <= 0 {
		query.Limit = 10 // Default limit
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	return h.paymentRepo.GetByUserID(ctx, query.UserID, query.Offset, query.Limit)
}