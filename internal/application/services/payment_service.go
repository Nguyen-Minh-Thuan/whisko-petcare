package services

import (
	"context"
	"fmt"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/payos"
)

// PaymentService handles payment business logic
type PaymentService struct {
	paymentRepo repository.PaymentRepository
	payOSClient *payos.PayOSClient
}

// NewPaymentService creates a new payment service
func NewPaymentService(paymentRepo repository.PaymentRepository, payOSClient *payos.PayOSClient) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		payOSClient: payOSClient,
	}
}

// CreatePaymentRequest represents a payment creation request
type CreatePaymentRequest struct {
	UserID      string                  `json:"user_id"`
	Amount      int                     `json:"amount"` // Amount in VND
	Description string                  `json:"description"`
	Items       []aggregate.PaymentItem `json:"items"`
}

// CreatePaymentResponse represents a payment creation response
type CreatePaymentResponse struct {
	PaymentID   string `json:"payment_id"`
	OrderCode   int64  `json:"order_code"`
	CheckoutURL string `json:"checkout_url"`
	QRCode      string `json:"qr_code"`
	Amount      int    `json:"amount"`
	Status      string `json:"status"`
	ExpiredAt   string `json:"expired_at"`
}

// CreatePayment creates a new payment and integrates with PayOS
func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Validate request
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	// Validate that item total matches amount
	totalAmount := 0
	for _, item := range req.Items {
		totalAmount += item.Price * item.Quantity
	}
	if totalAmount != req.Amount {
		return nil, fmt.Errorf("total item amount (%d) does not match payment amount (%d)", totalAmount, req.Amount)
	}

	// Create payment aggregate
	payment, err := aggregate.NewPayment(req.UserID, req.Amount, req.Description, req.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Convert items for PayOS API
	payOSItems := make([]payos.PaymentItem, len(req.Items))
	for i, item := range req.Items {
		payOSItems[i] = payos.PaymentItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	// Create payment request for PayOS
	payOSReq := &payos.CreatePaymentRequest{
		OrderCode:   payment.OrderCode(),
		Amount:      req.Amount,
		Description: req.Description,
		Items:       payOSItems,
	}

	// Create payment in PayOS
	payOSResp, err := s.payOSClient.CreatePayment(ctx, payOSReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create PayOS payment: %w", err)
	}

	if !payOSResp.Success {
		return nil, fmt.Errorf("PayOS payment creation failed: %s", payOSResp.Desc)
	}

	// Update payment with PayOS details
	err = payment.SetPayOSDetails(
		payOSResp.Data.PaymentLinkId,
		payOSResp.Data.CheckoutUrl,
		payOSResp.Data.QrCode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set PayOS details: %w", err)
	}

	// Save payment to repository
	err = s.paymentRepo.Save(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &CreatePaymentResponse{
		PaymentID:   payment.ID(),
		OrderCode:   payment.OrderCode(),
		CheckoutURL: payment.CheckoutURL(),
		QRCode:      payment.QRCode(),
		Amount:      payment.Amount(),
		Status:      string(payment.Status()),
		ExpiredAt:   payment.ExpiredAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetPayment retrieves a payment by ID
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*aggregate.Payment, error) {
	return s.paymentRepo.GetByID(ctx, paymentID)
}

// GetPaymentByOrderCode retrieves a payment by order code
func (s *PaymentService) GetPaymentByOrderCode(ctx context.Context, orderCode int64) (*aggregate.Payment, error) {
	return s.paymentRepo.GetByOrderCode(ctx, orderCode)
}

// GetUserPayments retrieves payments for a user
func (s *PaymentService) GetUserPayments(ctx context.Context, userID string, offset, limit int) ([]*aggregate.Payment, error) {
	return s.paymentRepo.GetByUserID(ctx, userID, offset, limit)
}

// ConfirmPayment confirms a payment (typically called from webhook)
func (s *PaymentService) ConfirmPayment(ctx context.Context, orderCode int64) error {
	// Get payment by order code
	payment, err := s.paymentRepo.GetByOrderCode(ctx, orderCode)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Verify payment status with PayOS
	payOSInfo, err := s.payOSClient.GetPaymentInfo(ctx, orderCode)
	if err != nil {
		return fmt.Errorf("failed to get PayOS payment info: %w", err)
	}

	if !payOSInfo.Success {
		return fmt.Errorf("PayOS payment info request failed: %s", payOSInfo.Desc)
	}

	// Update payment status based on PayOS response
	switch payOSInfo.Data.Status {
	case "PAID":
		err = payment.MarkAsPaid()
	case "CANCELLED":
		err = payment.MarkAsCancelled()
	case "EXPIRED":
		err = payment.MarkAsExpired()
	default:
		// Payment is still pending or in unknown state
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Save updated payment
	return s.paymentRepo.Save(ctx, payment)
}

// CancelPayment cancels a payment
func (s *PaymentService) CancelPayment(ctx context.Context, paymentID string, reason string) error {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Cancel in PayOS if still pending
	if payment.Status() == aggregate.PaymentStatusPending {
		err = s.payOSClient.CancelPayment(ctx, payment.OrderCode(), reason)
		if err != nil {
			return fmt.Errorf("failed to cancel PayOS payment: %w", err)
		}
	}

	// Mark payment as cancelled
	err = payment.MarkAsCancelled()
	if err != nil {
		return fmt.Errorf("failed to mark payment as cancelled: %w", err)
	}

	// Save updated payment
	return s.paymentRepo.Save(ctx, payment)
}

// ProcessWebhook processes PayOS webhook notifications
func (s *PaymentService) ProcessWebhook(ctx context.Context, webhookData *payos.WebhookData) error {
	// Confirm the payment based on webhook data
	return s.ConfirmPayment(ctx, webhookData.OrderCode)
}
