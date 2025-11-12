package command

import "whisko-petcare/internal/domain/aggregate"

// RequestPayoutCmd represents a vendor requesting a payout (auto-created after schedule creation)
type RequestPayoutCmd struct {
	PayoutID      string
	VendorID      string
	PaymentID     string // Link to the payment that triggered this
	ScheduleID    string // Link to the schedule that was created
	Amount        int
	BankName      string
	AccountNumber string
	AccountName   string
	BankBranch    string
	Notes         string
}

// ApprovePayoutCmd represents admin approving a payout request
type ApprovePayoutCmd struct {
	PayoutID string
	AdminID  string
	Notes    string
}

// RejectPayoutCmd represents admin rejecting a payout request
type RejectPayoutCmd struct {
	PayoutID string
	AdminID  string
	Reason   string
}

// ProcessPayoutCmd represents processing approved payout via PayOS
type ProcessPayoutCmd struct {
	PayoutID        string
	PayosTransferID string
}

// CompletePayoutCmd represents marking payout as completed (from webhook)
type CompletePayoutCmd struct {
	PayoutID string
}

// FailPayoutCmd represents marking payout as failed (from webhook or error)
type FailPayoutCmd struct {
	PayoutID string
	Reason   string
}

// UpdateVendorBankAccountCmd represents updating vendor's bank account
type UpdateVendorBankAccountCmd struct {
	VendorID      string
	BankName      string
	AccountNumber string
	AccountName   string
	BankBranch    string
}

// GetPayoutsByVendorQuery retrieves payouts for a vendor
type GetPayoutsByVendorQuery struct {
	VendorID string
	Offset   int
	Limit    int
}

// GetPayoutsByStatusQuery retrieves payouts by status
type GetPayoutsByStatusQuery struct {
	Status aggregate.PayoutStatus
	Offset int
	Limit  int
}

// GetPayoutByIDQuery retrieves a specific payout
type GetPayoutByIDQuery struct {
	PayoutID string
}
