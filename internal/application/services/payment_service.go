package services

import (
	"context"
	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/infrastructure/projection"
)

// PaymentService orchestrates payment operations
type PaymentService struct {
	// Command handlers (using Unit of Work)
	createPaymentHandler  *command.CreatePaymentWithUoWHandler
	cancelPaymentHandler  *command.CancelPaymentWithUoWHandler
	confirmPaymentHandler *command.ConfirmPaymentWithUoWHandler

	// Query handlers (using Projections)
	getPaymentHandler            *query.GetPaymentHandler
	getPaymentByOrderCodeHandler *query.GetPaymentByOrderCodeHandler
	listUserPaymentsHandler      *query.ListUserPaymentsHandler
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	createPaymentHandler *command.CreatePaymentWithUoWHandler,
	cancelPaymentHandler *command.CancelPaymentWithUoWHandler,
	confirmPaymentHandler *command.ConfirmPaymentWithUoWHandler,
	getPaymentHandler *query.GetPaymentHandler,
	getPaymentByOrderCodeHandler *query.GetPaymentByOrderCodeHandler,
	listUserPaymentsHandler *query.ListUserPaymentsHandler,
) *PaymentService {
	return &PaymentService{
		createPaymentHandler:         createPaymentHandler,
		cancelPaymentHandler:         cancelPaymentHandler,
		confirmPaymentHandler:        confirmPaymentHandler,
		getPaymentHandler:            getPaymentHandler,
		getPaymentByOrderCodeHandler: getPaymentByOrderCodeHandler,
		listUserPaymentsHandler:      listUserPaymentsHandler,
	}
}

// Command operations

// CreatePayment creates a new payment with PayOS integration
func (s *PaymentService) CreatePayment(ctx context.Context, cmd command.CreatePaymentCommand) (*command.CreatePaymentResponse, error) {
	return s.createPaymentHandler.Handle(ctx, &cmd)
}

// CancelPayment cancels a payment
func (s *PaymentService) CancelPayment(ctx context.Context, cmd command.CancelPaymentCommand) error {
	return s.cancelPaymentHandler.Handle(ctx, &cmd)
}

// ConfirmPayment confirms a payment (typically called from webhook)
func (s *PaymentService) ConfirmPayment(ctx context.Context, cmd command.ConfirmPaymentCommand) error {
	return s.confirmPaymentHandler.Handle(ctx, &cmd)
}

// Query operations

// GetPayment retrieves a payment by ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*projection.PaymentReadModel, error) {
	return s.getPaymentHandler.Handle(ctx, &query.GetPaymentQuery{PaymentID: paymentID})
}

// GetPaymentByOrderCode retrieves a payment by order code
func (s *PaymentService) GetPaymentByOrderCode(ctx context.Context, orderCode int64) (*projection.PaymentReadModel, error) {
	return s.getPaymentByOrderCodeHandler.Handle(ctx, &query.GetPaymentByOrderCodeQuery{OrderCode: orderCode})
}

// GetUserPayments retrieves payments for a user
func (s *PaymentService) GetUserPayments(ctx context.Context, userID string, offset, limit int) ([]*projection.PaymentReadModel, error) {
	return s.listUserPaymentsHandler.Handle(ctx, &query.ListUserPaymentsQuery{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})
}
