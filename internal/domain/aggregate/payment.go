package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusPaid      PaymentStatus = "PAID"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusExpired   PaymentStatus = "EXPIRED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodPayOS PaymentMethod = "PAYOS"
)

// PaymentItem represents an item in the payment
type PaymentItem = event.PaymentItem

// Payment represents a payment aggregate root
type Payment struct {
	id                string
	orderCode         int64
	userID            string
	amount            int // Amount in VND cents
	description       string
	items             []PaymentItem
	status            PaymentStatus
	method            PaymentMethod
	payOSTransactionID string
	checkoutURL       string
	qrCode            string
	expiredAt         time.Time
	version           int
	createdAt         time.Time
	updatedAt         time.Time
	uncommittedEvents []event.DomainEvent
}

// NewPayment creates a new payment aggregate
func NewPayment(userID string, amount int, description string, items []PaymentItem) (*Payment, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items cannot be empty")
	}

	// Generate unique order code (timestamp + random)
	orderCode := time.Now().Unix()*1000 + int64(time.Now().Nanosecond()/1000000)
	
	payment := &Payment{
		id:          uuid.New().String(),
		orderCode:   orderCode,
		userID:      userID,
		amount:      amount,
		description: description,
		items:       items,
		status:      PaymentStatusPending,
		method:      PaymentMethodPayOS,
		expiredAt:   time.Now().Add(15 * time.Minute), // Default 15 minutes expiration
		version:     1,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}

	payment.raiseEvent(&event.PaymentCreated{
		PaymentID:   payment.id,
		UserID:      userID,
		OrderCode:   orderCode,
		Amount:      amount,
		Description: description,
		Items:       items,
		Status:      string(PaymentStatusPending),
		Method:      string(PaymentMethodPayOS),
		ExpiredAt:   payment.expiredAt,
		Timestamp:   payment.createdAt,
	})

	return payment, nil
}

// SetPayOSDetails updates the payment with PayOS transaction details
func (p *Payment) SetPayOSDetails(transactionID, checkoutURL, qrCode string) error {
	if p.status != PaymentStatusPending {
		return fmt.Errorf("cannot set PayOS details for payment with status: %s", p.status)
	}

	p.payOSTransactionID = transactionID
	p.checkoutURL = checkoutURL
	p.qrCode = qrCode
	p.version++
	p.updatedAt = time.Now()

	p.raiseEvent(&event.PaymentUpdated{
		PaymentID:          p.id,
		PayOSTransactionID: transactionID,
		CheckoutURL:        checkoutURL,
		QRCode:             qrCode,
		Timestamp:          p.updatedAt,
	})

	return nil
}

// MarkAsPaid marks the payment as paid
func (p *Payment) MarkAsPaid() error {
	if p.status != PaymentStatusPending {
		return fmt.Errorf("cannot mark payment as paid with current status: %s", p.status)
	}

	p.status = PaymentStatusPaid
	p.version++
	p.updatedAt = time.Now()

	p.raiseEvent(&event.PaymentStatusChanged{
		PaymentID: p.id,
		OldStatus: string(PaymentStatusPending),
		NewStatus: string(PaymentStatusPaid),
		Timestamp: p.updatedAt,
	})

	return nil
}

// MarkAsCancelled marks the payment as cancelled
func (p *Payment) MarkAsCancelled() error {
	if p.status == PaymentStatusPaid {
		return fmt.Errorf("cannot cancel a paid payment")
	}

	oldStatus := p.status
	p.status = PaymentStatusCancelled
	p.version++
	p.updatedAt = time.Now()

	p.raiseEvent(&event.PaymentStatusChanged{
		PaymentID: p.id,
		OldStatus: string(oldStatus),
		NewStatus: string(PaymentStatusCancelled),
		Timestamp: p.updatedAt,
	})

	return nil
}

// MarkAsExpired marks the payment as expired
func (p *Payment) MarkAsExpired() error {
	if p.status == PaymentStatusPaid {
		return fmt.Errorf("cannot expire a paid payment")
	}

	oldStatus := p.status
	p.status = PaymentStatusExpired
	p.version++
	p.updatedAt = time.Now()

	p.raiseEvent(&event.PaymentStatusChanged{
		PaymentID: p.id,
		OldStatus: string(oldStatus),
		NewStatus: string(PaymentStatusExpired),
		Timestamp: p.updatedAt,
	})

	return nil
}

// MarkAsFailed marks the payment as failed
func (p *Payment) MarkAsFailed() error {
	if p.status == PaymentStatusPaid {
		return fmt.Errorf("cannot fail a paid payment")
	}

	oldStatus := p.status
	p.status = PaymentStatusFailed
	p.version++
	p.updatedAt = time.Now()

	p.raiseEvent(&event.PaymentStatusChanged{
		PaymentID: p.id,
		OldStatus: string(oldStatus),
		NewStatus: string(PaymentStatusFailed),
		Timestamp: p.updatedAt,
	})

	return nil
}

// IsExpired checks if the payment has expired
func (p *Payment) IsExpired() bool {
	return time.Now().After(p.expiredAt) && p.status == PaymentStatusPending
}

// raiseEvent adds an event to the uncommitted events
func (p *Payment) raiseEvent(evt event.DomainEvent) {
	p.uncommittedEvents = append(p.uncommittedEvents, evt)
}

// applyEvent applies an event to the payment state
func (p *Payment) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.PaymentCreated:
		p.id = e.PaymentID
		p.userID = e.UserID
		p.orderCode = e.OrderCode
		p.amount = e.Amount
		p.description = e.Description
		p.items = e.Items
		p.status = PaymentStatus(e.Status)
		p.method = PaymentMethod(e.Method)
		p.expiredAt = e.ExpiredAt
		p.createdAt = e.Timestamp
		p.updatedAt = e.Timestamp

	case *event.PaymentUpdated:
		p.payOSTransactionID = e.PayOSTransactionID
		p.checkoutURL = e.CheckoutURL
		p.qrCode = e.QRCode
		p.updatedAt = e.Timestamp

	case *event.PaymentStatusChanged:
		p.status = PaymentStatus(e.NewStatus)
		p.updatedAt = e.Timestamp

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}

	p.version++
	return nil
}

// LoadFromHistory reconstructs the payment from its event history
func (p *Payment) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := p.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}

// Getters
func (p *Payment) ID() string                        { return p.id }
func (p *Payment) OrderCode() int64                  { return p.orderCode }
func (p *Payment) UserID() string                    { return p.userID }
func (p *Payment) Amount() int                       { return p.amount }
func (p *Payment) Description() string               { return p.description }
func (p *Payment) Items() []PaymentItem              { return p.items }
func (p *Payment) Status() PaymentStatus             { return p.status }
func (p *Payment) Method() PaymentMethod             { return p.method }
func (p *Payment) PayOSTransactionID() string        { return p.payOSTransactionID }
func (p *Payment) CheckoutURL() string               { return p.checkoutURL }
func (p *Payment) QRCode() string                    { return p.qrCode }
func (p *Payment) ExpiredAt() time.Time              { return p.expiredAt }
func (p *Payment) Version() int                      { return p.version }
func (p *Payment) CreatedAt() time.Time              { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time              { return p.updatedAt }

// Entity interface implementation
func (p *Payment) GetID() string    { return p.id }
func (p *Payment) SetID(id string)  { p.id = id }
func (p *Payment) GetVersion() int  { return p.version }
func (p *Payment) SetVersion(v int) { p.version = v }

// AggregateRoot interface implementation
func (p *Payment) GetUncommittedEvents() []event.DomainEvent {
	return p.uncommittedEvents
}

func (p *Payment) MarkEventsAsCommitted() {
	p.uncommittedEvents = nil
}