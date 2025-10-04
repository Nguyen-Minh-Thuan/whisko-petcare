package command

import (
	"context"
	"fmt"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/payos"
)

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

// CreatePaymentCommand represents a command to create a new payment
type CreatePaymentCommand struct {
	UserID      string                    `json:"user_id"`
	Amount      int                       `json:"amount"`      // Amount in VND
	Description string                    `json:"description"`
	Items       []aggregate.PaymentItem   `json:"items"`
}

// CreatePaymentHandler handles create payment commands
type CreatePaymentHandler struct {
	paymentRepo  repository.PaymentRepository
	payOSService *payos.Service
}

// NewCreatePaymentHandler creates a new create payment handler
func NewCreatePaymentHandler(paymentRepo repository.PaymentRepository, payOSService *payos.Service) *CreatePaymentHandler {
	return &CreatePaymentHandler{
		paymentRepo:  paymentRepo,
		payOSService: payOSService,
	}
}

// Handle processes the create payment command
func (h *CreatePaymentHandler) Handle(ctx context.Context, cmd *CreatePaymentCommand) (*CreatePaymentResponse, error) {
	if cmd == nil {
		return nil, fmt.Errorf("command cannot be nil")
	}

	// Validate request
	if cmd.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if cmd.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if cmd.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if len(cmd.Items) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	// Validate that item total matches amount
	totalAmount := 0
	for _, item := range cmd.Items {
		totalAmount += item.Price * item.Quantity
	}
	if totalAmount != cmd.Amount {
		return nil, fmt.Errorf("total item amount (%d) does not match payment amount (%d)", totalAmount, cmd.Amount)
	}

	// Create payment aggregate
	payment, err := aggregate.NewPayment(cmd.UserID, cmd.Amount, cmd.Description, cmd.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Convert items for PayOS API
	payOSItems := make([]payos.PaymentItem, len(cmd.Items))
	for i, item := range cmd.Items {
		payOSItems[i] = payos.PaymentItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	// Create payment request for PayOS
	payOSReq := &payos.CreatePaymentRequest{
		OrderCode:   payment.OrderCode(),
		Amount:      cmd.Amount,
		Description: cmd.Description,
		Items:       payOSItems,
	}

	// Create payment in PayOS using the new service
	payOSResp, err := h.payOSService.CreatePaymentLink(ctx, payOSReq)
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
	err = h.paymentRepo.Save(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &CreatePaymentResponse{
		PaymentID:   payment.ID(),
		OrderCode:   payment.OrderCode(),
		CheckoutURL: payOSResp.Data.CheckoutUrl,
		QRCode:      payOSResp.Data.QrCode,
		Amount:      payment.Amount(),
		Status:      string(payment.Status()),
		ExpiredAt:   payment.ExpiredAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CancelPaymentCommand represents a command to cancel a payment
type CancelPaymentCommand struct {
	PaymentID string `json:"payment_id"`
	Reason    string `json:"reason"`
}

// CancelPaymentHandler handles cancel payment commands
type CancelPaymentHandler struct {
	paymentRepo  repository.PaymentRepository
	payOSService *payos.Service
}

// NewCancelPaymentHandler creates a new cancel payment handler
func NewCancelPaymentHandler(paymentRepo repository.PaymentRepository, payOSService *payos.Service) *CancelPaymentHandler {
	return &CancelPaymentHandler{
		paymentRepo:  paymentRepo,
		payOSService: payOSService,
	}
}

// Handle processes the cancel payment command
func (h *CancelPaymentHandler) Handle(ctx context.Context, cmd *CancelPaymentCommand) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}

	if cmd.PaymentID == "" {
		return fmt.Errorf("payment_id is required")
	}

	// Get payment
	payment, err := h.paymentRepo.GetByID(ctx, cmd.PaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	reason := cmd.Reason
	if reason == "" {
		reason = "Cancelled by user"
	}

	// Cancel in PayOS if still pending
	if payment.Status() == aggregate.PaymentStatusPending {
		err = h.payOSService.CancelPaymentLink(ctx, payment.OrderCode(), reason)
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
	return h.paymentRepo.Save(ctx, payment)
}

// ConfirmPaymentCommand represents a command to confirm a payment (typically from webhook)
type ConfirmPaymentCommand struct {
	OrderCode int64 `json:"order_code"`
}

// ConfirmPaymentHandler handles confirm payment commands
type ConfirmPaymentHandler struct {
	paymentRepo  repository.PaymentRepository
	payOSService *payos.Service
}

// NewConfirmPaymentHandler creates a new confirm payment handler
func NewConfirmPaymentHandler(paymentRepo repository.PaymentRepository, payOSService *payos.Service) *ConfirmPaymentHandler {
	return &ConfirmPaymentHandler{
		paymentRepo:  paymentRepo,
		payOSService: payOSService,
	}
}

// Handle processes the confirm payment command
func (h *ConfirmPaymentHandler) Handle(ctx context.Context, cmd *ConfirmPaymentCommand) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}

	if cmd.OrderCode == 0 {
		return fmt.Errorf("order_code is required")
	}

	// Get payment by order code
	payment, err := h.paymentRepo.GetByOrderCode(ctx, cmd.OrderCode)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Verify payment status with PayOS
	payOSInfo, err := h.payOSService.GetPaymentLinkInformation(ctx, cmd.OrderCode)
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
	return h.paymentRepo.Save(ctx, payment)
}