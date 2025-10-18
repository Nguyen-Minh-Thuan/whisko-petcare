package event

import "time"

// PaymentItem represents an item in the payment (duplicate from aggregate to avoid circular dependency)
type PaymentItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"` // Amount in VND cents
}

// PaymentCreated event
type PaymentCreated struct {
	PaymentID   string        `json:"payment_id"`
	UserID      string        `json:"user_id"`
	OrderCode   int64         `json:"order_code"`
	Amount      int           `json:"amount"`
	Description string        `json:"description"`
	Items       []PaymentItem `json:"items"`
	Status      string        `json:"status"`
	Method      string        `json:"method"`
	ExpiredAt   time.Time     `json:"expired_at"`
	Timestamp   time.Time     `json:"timestamp"`
}

func (e *PaymentCreated) EventType() string     { return "PaymentCreated" }
func (e *PaymentCreated) AggregateID() string   { return e.PaymentID }
func (e *PaymentCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *PaymentCreated) Version() int          { return 1 }

// PaymentUpdated event
type PaymentUpdated struct {
	PaymentID          string    `json:"payment_id"`
	PayOSTransactionID string    `json:"payos_transaction_id"`
	CheckoutURL        string    `json:"checkout_url"`
	QRCode             string    `json:"qr_code"`
	Timestamp          time.Time `json:"timestamp"`
}

func (e *PaymentUpdated) EventType() string     { return "PaymentUpdated" }
func (e *PaymentUpdated) AggregateID() string   { return e.PaymentID }
func (e *PaymentUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *PaymentUpdated) Version() int          { return 1 }

// PaymentStatusChanged event
type PaymentStatusChanged struct {
	PaymentID string    `json:"payment_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PaymentStatusChanged) EventType() string     { return "PaymentStatusChanged" }
func (e *PaymentStatusChanged) AggregateID() string   { return e.PaymentID }
func (e *PaymentStatusChanged) OccurredAt() time.Time { return e.Timestamp }
func (e *PaymentStatusChanged) Version() int          { return 1 }