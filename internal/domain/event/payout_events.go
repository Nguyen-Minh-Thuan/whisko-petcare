package event

import "time"

// PayoutRequested event - fired when vendor requests a payout
type PayoutRequested struct {
	PayoutID      string    `json:"payout_id"`
	VendorID      string    `json:"vendor_id"`
	PaymentID     string    `json:"payment_id"`  // Link to the payment that triggered this
	ScheduleID    string    `json:"schedule_id"` // Link to the schedule that was created
	Amount        int       `json:"amount"`
	BankName      string    `json:"bank_name"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	BankBranch    string    `json:"bank_branch"`
	Notes         string    `json:"notes"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PayoutRequested) EventType() string     { return "PayoutRequested" }
func (e *PayoutRequested) AggregateID() string   { return e.PayoutID }
func (e *PayoutRequested) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutRequested) Version() int          { return 1 }

// PayoutApproved event - fired when admin approves a payout
type PayoutApproved struct {
	PayoutID     string    `json:"payout_id"`
	VendorID     string    `json:"vendor_id"`
	Amount       int       `json:"amount"`
	ApprovedBy   string    `json:"approved_by"` // Admin ID
	ApprovedAt   time.Time `json:"approved_at"`
	Notes        string    `json:"notes"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PayoutApproved) EventType() string     { return "PayoutApproved" }
func (e *PayoutApproved) AggregateID() string   { return e.PayoutID }
func (e *PayoutApproved) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutApproved) Version() int          { return e.EventVersion }

// PayoutRejected event - fired when admin rejects a payout
type PayoutRejected struct {
	PayoutID     string    `json:"payout_id"`
	VendorID     string    `json:"vendor_id"`
	Amount       int       `json:"amount"`
	RejectedBy   string    `json:"rejected_by"` // Admin ID
	RejectedAt   time.Time `json:"rejected_at"`
	Reason       string    `json:"reason"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PayoutRejected) EventType() string     { return "PayoutRejected" }
func (e *PayoutRejected) AggregateID() string   { return e.PayoutID }
func (e *PayoutRejected) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutRejected) Version() int          { return e.EventVersion }

// PayoutProcessing event - fired when payout is sent to PayOS
type PayoutProcessing struct {
	PayoutID        string    `json:"payout_id"`
	VendorID        string    `json:"vendor_id"`
	Amount          int       `json:"amount"`
	PayosTransferID string    `json:"payos_transfer_id"`
	ProcessedAt     time.Time `json:"processed_at"`
	EventVersion    int       `json:"version"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e *PayoutProcessing) EventType() string     { return "PayoutProcessing" }
func (e *PayoutProcessing) AggregateID() string   { return e.PayoutID }
func (e *PayoutProcessing) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutProcessing) Version() int          { return e.EventVersion }

// PayoutCompleted event - fired when PayOS confirms transfer success
type PayoutCompleted struct {
	PayoutID        string    `json:"payout_id"`
	VendorID        string    `json:"vendor_id"`
	Amount          int       `json:"amount"`
	PayosTransferID string    `json:"payos_transfer_id"`
	CompletedAt     time.Time `json:"completed_at"`
	EventVersion    int       `json:"version"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e *PayoutCompleted) EventType() string     { return "PayoutCompleted" }
func (e *PayoutCompleted) AggregateID() string   { return e.PayoutID }
func (e *PayoutCompleted) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutCompleted) Version() int          { return e.EventVersion }

// PayoutFailed event - fired when PayOS transfer fails
type PayoutFailed struct {
	PayoutID     string    `json:"payout_id"`
	VendorID     string    `json:"vendor_id"`
	Amount       int       `json:"amount"`
	Reason       string    `json:"reason"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PayoutFailed) EventType() string     { return "PayoutFailed" }
func (e *PayoutFailed) AggregateID() string   { return e.PayoutID }
func (e *PayoutFailed) OccurredAt() time.Time { return e.Timestamp }
func (e *PayoutFailed) Version() int          { return e.EventVersion }

// VendorBankAccountUpdated event - fired when vendor updates bank account info
type VendorBankAccountUpdated struct {
	VendorID      string    `json:"vendor_id"`
	BankName      string    `json:"bank_name"`
	AccountNumber string    `json:"account_number"`
	AccountName   string    `json:"account_name"`
	BankBranch    string    `json:"bank_branch"`
	EventVersion  int       `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *VendorBankAccountUpdated) EventType() string     { return "VendorBankAccountUpdated" }
func (e *VendorBankAccountUpdated) AggregateID() string   { return e.VendorID }
func (e *VendorBankAccountUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorBankAccountUpdated) Version() int          { return e.EventVersion }
