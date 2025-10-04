package event

import "time"

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	Version() int
}

// UserCreated event
type UserCreated struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *UserCreated) EventType() string     { return "UserCreated" }
func (e *UserCreated) AggregateID() string   { return e.UserID }
func (e *UserCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserCreated) Version() int          { return 1 }

// UserProfileUpdated event
type UserProfileUpdated struct {
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserProfileUpdated) EventType() string     { return "UserProfileUpdated" }
func (e *UserProfileUpdated) AggregateID() string   { return e.UserID }
func (e *UserProfileUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserProfileUpdated) Version() int          { return e.EventVersion }

// UserContactUpdated event
type UserContactUpdated struct {
	UserID       string    `json:"user_id"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserContactUpdated) EventType() string     { return "UserContactUpdated" }
func (e *UserContactUpdated) AggregateID() string   { return e.UserID }
func (e *UserContactUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserContactUpdated) Version() int          { return e.EventVersion }

// UserDeleted event
type UserDeleted struct {
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *UserDeleted) EventType() string     { return "UserDeleted" }
func (e *UserDeleted) AggregateID() string   { return e.UserID }
func (e *UserDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *UserDeleted) Version() int          { return 0 }

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
