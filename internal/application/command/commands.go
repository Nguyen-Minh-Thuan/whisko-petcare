package command

import "whisko-petcare/internal/domain/aggregate"

// ============================================
// User Commands
// ============================================

// CreateUser represents a command to create a new user
type CreateUser struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

// UpdateUserProfile represents a command to update user profile
type UpdateUserProfile struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// UpdateUserContact represents a command to update user contact info
type UpdateUserContact struct {
	UserID  string `json:"user_id"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// DeleteUser represents a command to delete a user
type DeleteUser struct {
	UserID string `json:"user_id"`
}

// ============================================
// Payment Commands
// ============================================

// CreatePaymentCommand represents a command to create a new payment
type CreatePaymentCommand struct {
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

// CancelPaymentCommand represents a command to cancel a payment
type CancelPaymentCommand struct {
	PaymentID string `json:"payment_id"`
	Reason    string `json:"reason"`
}

// ConfirmPaymentCommand represents a command to confirm a payment (typically from webhook)
type ConfirmPaymentCommand struct {
	OrderCode int64 `json:"order_code"`
}
