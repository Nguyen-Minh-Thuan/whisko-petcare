package query

import (
	"context"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/errors"
)

// GetPaymentQuery represents a query to get a payment by ID
type GetPaymentQuery struct {
	PaymentID string `json:"payment_id"`
}

// GetPaymentHandler handles get payment queries
type GetPaymentHandler struct {
	paymentProjection projection.PaymentProjection
}

// NewGetPaymentHandler creates a new get payment handler
func NewGetPaymentHandler(paymentProjection projection.PaymentProjection) *GetPaymentHandler {
	return &GetPaymentHandler{
		paymentProjection: paymentProjection,
	}
}

// Handle processes the get payment query
func (h *GetPaymentHandler) Handle(ctx context.Context, query *GetPaymentQuery) (*projection.PaymentReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	if query.PaymentID == "" {
		return nil, errors.NewValidationError("payment_id is required")
	}

	payment, err := h.paymentProjection.GetByID(ctx, query.PaymentID)
	if err != nil {
		return nil, errors.NewNotFoundError("payment")
	}

	return payment, nil
}

// GetPaymentByOrderCodeQuery represents a query to get a payment by order code
type GetPaymentByOrderCodeQuery struct {
	OrderCode int64 `json:"order_code"`
}

// GetPaymentByOrderCodeHandler handles get payment by order code queries
type GetPaymentByOrderCodeHandler struct {
	paymentProjection projection.PaymentProjection
}

// NewGetPaymentByOrderCodeHandler creates a new get payment by order code handler
func NewGetPaymentByOrderCodeHandler(paymentProjection projection.PaymentProjection) *GetPaymentByOrderCodeHandler {
	return &GetPaymentByOrderCodeHandler{
		paymentProjection: paymentProjection,
	}
}

// Handle processes the get payment by order code query
func (h *GetPaymentByOrderCodeHandler) Handle(ctx context.Context, query *GetPaymentByOrderCodeQuery) (*projection.PaymentReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	if query.OrderCode == 0 {
		return nil, errors.NewValidationError("order_code is required")
	}

	payment, err := h.paymentProjection.GetByOrderCode(ctx, query.OrderCode)
	if err != nil {
		return nil, errors.NewNotFoundError("payment")
	}

	return payment, nil
}

// ListUserPaymentsQuery represents a query to list payments for a user
type ListUserPaymentsQuery struct {
	UserID string `json:"user_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// ListUserPaymentsHandler handles list user payments queries
type ListUserPaymentsHandler struct {
	paymentProjection projection.PaymentProjection
}

// NewListUserPaymentsHandler creates a new list user payments handler
func NewListUserPaymentsHandler(paymentProjection projection.PaymentProjection) *ListUserPaymentsHandler {
	return &ListUserPaymentsHandler{
		paymentProjection: paymentProjection,
	}
}

// Handle processes the list user payments query
func (h *ListUserPaymentsHandler) Handle(ctx context.Context, query *ListUserPaymentsQuery) ([]*projection.PaymentReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	if query.UserID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}

	if query.Limit <= 0 {
		query.Limit = 10 // Default limit
	}
	if query.Limit > 100 {
		query.Limit = 100 // Max limit
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	payments, err := h.paymentProjection.ListByUserID(ctx, query.UserID, query.Limit, query.Offset)
	if err != nil {
		return nil, errors.NewInternalError("failed to list payments")
	}

	return payments, nil
}

