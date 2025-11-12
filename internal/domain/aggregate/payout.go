package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"
)

// PayoutStatus represents the status of a payout request
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "PENDING"     // Auto-created, waiting to process
	PayoutStatusProcessing PayoutStatus = "PROCESSING"  // Sent to PayOS, waiting for completion
	PayoutStatusCompleted  PayoutStatus = "COMPLETED"   // Successfully transferred
	PayoutStatusFailed     PayoutStatus = "FAILED"      // PayOS transfer failed
)

// BankAccount represents vendor's bank account information
type BankAccount struct {
	BankName      string
	AccountNumber string
	AccountName   string
	BankBranch    string
}

// Payout represents a payout request from vendor
type Payout struct {
	id              string
	vendorID        string
	paymentID       string // Link to the payment that triggered this payout
	scheduleID      string // Link to the schedule that was created
	amount          int
	status          PayoutStatus
	requestedAt     time.Time
	processedAt     *time.Time
	completedAt     *time.Time
	payosTransferID string
	notes           string
	failureReason   string
	bankAccount     BankAccount
	version         int
	createdAt       time.Time
	updatedAt       time.Time

	uncommittedEvents []event.DomainEvent
}

// NewPayout creates a new payout request (auto-created after schedule is created from payment)
func NewPayout(payoutID, vendorID, paymentID, scheduleID string, amount int, bankAccount BankAccount, notes string) (*Payout, error) {
	if payoutID == "" {
		return nil, fmt.Errorf("payout ID cannot be empty")
	}
	if vendorID == "" {
		return nil, fmt.Errorf("vendor ID cannot be empty")
	}
	if paymentID == "" {
		return nil, fmt.Errorf("payment ID cannot be empty")
	}
	if scheduleID == "" {
		return nil, fmt.Errorf("schedule ID cannot be empty")
	}
	if amount < 100000 {
		return nil, fmt.Errorf("minimum payout amount is 100,000 VND")
	}
	if amount > 50000000 {
		return nil, fmt.Errorf("maximum payout amount is 50,000,000 VND per transaction")
	}
	if bankAccount.BankName == "" || bankAccount.AccountNumber == "" || bankAccount.AccountName == "" {
		return nil, fmt.Errorf("complete bank account information is required")
	}

	now := time.Now()
	payout := &Payout{
		id:          payoutID,
		vendorID:    vendorID,
		paymentID:   paymentID,
		scheduleID:  scheduleID,
		amount:      amount,
		status:      PayoutStatusPending,
		requestedAt: now,
		bankAccount: bankAccount,
		notes:       notes,
		version:     1,
		createdAt:   now,
		updatedAt:   now,
	}

	payout.raiseEvent(&event.PayoutRequested{
		PayoutID:      payoutID,
		VendorID:      vendorID,
		PaymentID:     paymentID,
		ScheduleID:    scheduleID,
		Amount:        amount,
		BankName:      bankAccount.BankName,
		AccountNumber: bankAccount.AccountNumber,
		AccountName:   bankAccount.AccountName,
		BankBranch:    bankAccount.BankBranch,
		Notes:         notes,
		Timestamp:     now,
	})

	return payout, nil
}

// ReconstructPayout reconstructs a payout from database state (for MongoDB repository)
func ReconstructPayout(
	id, vendorID, paymentID, scheduleID string,
	amount int,
	bankAccount BankAccount,
	status, notes, failureReason string,
	version int,
	createdAt, updatedAt time.Time,
) *Payout {
	return &Payout{
		id:            id,
		vendorID:      vendorID,
		paymentID:     paymentID,
		scheduleID:    scheduleID,
		amount:        amount,
		status:        PayoutStatus(status),
		bankAccount:   bankAccount,
		notes:         notes,
		failureReason: failureReason,
		version:       version,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// MarkAsProcessing marks payout as being processed by PayOS (automatic)
func (p *Payout) MarkAsProcessing(payosTransferID string) error {
	if p.status != PayoutStatusPending {
		return fmt.Errorf("only pending payouts can be processed (current status: %s)", p.status)
	}
	if payosTransferID == "" {
		return fmt.Errorf("PayOS transfer ID is required")
	}

	now := time.Now()
	p.status = PayoutStatusProcessing
	p.processedAt = &now
	p.payosTransferID = payosTransferID
	p.version++
	p.updatedAt = now

	p.raiseEvent(&event.PayoutProcessing{
		PayoutID:        p.id,
		VendorID:        p.vendorID,
		Amount:          p.amount,
		PayosTransferID: payosTransferID,
		ProcessedAt:     now,
		EventVersion:    p.version,
		Timestamp:       now,
	})

	return nil
}

// MarkAsCompleted marks payout as completed (from webhook)
func (p *Payout) MarkAsCompleted() error {
	if p.status != PayoutStatusProcessing {
		return fmt.Errorf("only processing payouts can be completed (current status: %s)", p.status)
	}

	now := time.Now()
	p.status = PayoutStatusCompleted
	p.completedAt = &now
	p.version++
	p.updatedAt = now

	p.raiseEvent(&event.PayoutCompleted{
		PayoutID:        p.id,
		VendorID:        p.vendorID,
		Amount:          p.amount,
		PayosTransferID: p.payosTransferID,
		CompletedAt:     now,
		EventVersion:    p.version,
		Timestamp:       now,
	})

	return nil
}

// MarkAsFailed marks payout as failed (from webhook or error)
func (p *Payout) MarkAsFailed(reason string) error {
	// allow failing when processing or pending (e.g., immediate failure)
	if p.status != PayoutStatusProcessing && p.status != PayoutStatusPending {
		return fmt.Errorf("only processing or pending payouts can be failed (current status: %s)", p.status)
	}

	now := time.Now()
	p.status = PayoutStatusFailed
	p.failureReason = reason
	p.version++
	p.updatedAt = now

	p.raiseEvent(&event.PayoutFailed{
		PayoutID:     p.id,
		VendorID:     p.vendorID,
		Amount:       p.amount,
		Reason:       reason,
		EventVersion: p.version,
		Timestamp:    now,
	})

	return nil
}

func (p *Payout) raiseEvent(ev event.DomainEvent) {
	p.uncommittedEvents = append(p.uncommittedEvents, ev)
}

func (p *Payout) GetUncommittedEvents() []event.DomainEvent {
	return p.uncommittedEvents
}

func (p *Payout) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.PayoutRequested:
		p.id = e.PayoutID
		p.vendorID = e.VendorID
		p.paymentID = e.PaymentID
		p.scheduleID = e.ScheduleID
		p.amount = e.Amount
		p.status = PayoutStatusPending
		p.requestedAt = e.Timestamp
		p.bankAccount = BankAccount{
			BankName:      e.BankName,
			AccountNumber: e.AccountNumber,
			AccountName:   e.AccountName,
			BankBranch:    e.BankBranch,
		}
		p.notes = e.Notes
		p.createdAt = e.Timestamp
		p.updatedAt = e.Timestamp
		p.version = 1

	case *event.PayoutProcessing:
		p.status = PayoutStatusProcessing
		p.processedAt = &e.ProcessedAt
		p.payosTransferID = e.PayosTransferID
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp

	case *event.PayoutCompleted:
		p.status = PayoutStatusCompleted
		p.completedAt = &e.CompletedAt
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp

	case *event.PayoutFailed:
		p.status = PayoutStatusFailed
		p.failureReason = e.Reason
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}

	return nil
}

// Getters
func (p *Payout) ID() string               { return p.id }
func (p *Payout) VendorID() string         { return p.vendorID }
func (p *Payout) PaymentID() string        { return p.paymentID }
func (p *Payout) ScheduleID() string       { return p.scheduleID }
func (p *Payout) Amount() int              { return p.amount }
func (p *Payout) Status() PayoutStatus     { return p.status }
func (p *Payout) RequestedAt() time.Time   { return p.requestedAt }
func (p *Payout) ProcessedAt() *time.Time  { return p.processedAt }
func (p *Payout) CompletedAt() *time.Time  { return p.completedAt }
func (p *Payout) PayosTransferID() string  { return p.payosTransferID }
func (p *Payout) Notes() string            { return p.notes }
func (p *Payout) FailureReason() string    { return p.failureReason }
func (p *Payout) BankAccount() BankAccount { return p.bankAccount }
func (p *Payout) Version() int             { return p.version }
func (p *Payout) CreatedAt() time.Time     { return p.createdAt }
func (p *Payout) UpdatedAt() time.Time     { return p.updatedAt }

// Entity interface implementation
func (p *Payout) GetID() string          { return p.id }
func (p *Payout) GetVersion() int        { return p.version }
func (p *Payout) SetVersion(ver int)     { p.version = ver }

// AggregateRoot interface implementation
func (p *Payout) MarkEventsAsCommitted() {
	p.uncommittedEvents = nil
}

func (p *Payout) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := p.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}

func NewPayoutFromHistory(events []event.DomainEvent) (*Payout, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	payout := &Payout{}
	for _, ev := range events {
		if err := payout.applyEvent(ev); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", ev.EventType(), err)
		}
	}
	return payout, nil
}
