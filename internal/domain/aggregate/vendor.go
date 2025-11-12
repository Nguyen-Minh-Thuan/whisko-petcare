package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"
)

// VendorBankAccount holds vendor's bank information for payouts
type VendorBankAccount struct {
	BankName      string
	AccountNumber string
	AccountName   string
	BankBranch    string
}

type Vendor struct {
	id          string
	name        string
	email       string
	phone       string
	address     string
	imageUrl    string
	bankAccount *VendorBankAccount // Optional bank account for payouts
	version     int
	createdAt   time.Time
	updatedAt   time.Time
	isActive    bool

	uncommittedEvents []event.DomainEvent
}

func NewVendor(vendorID, name, email, phone, address string, imageUrl ...string) (*Vendor, error) {
	if vendorID == "" {
		return nil, fmt.Errorf("vendor ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if phone == "" {
		return nil, fmt.Errorf("phone cannot be empty")
	}
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	vendor := &Vendor{
		id:        vendorID,
		name:      name,
		email:     email,
		phone:     phone,
		address:   address,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		version:   1,
		isActive:  true,
	}

	// Set imageUrl if provided
	if len(imageUrl) > 0 && imageUrl[0] != "" {
		vendor.imageUrl = imageUrl[0]
	}

	vendor.raiseEvent(&event.VendorCreated{
		VendorID:  vendor.id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		Address:   address,
		ImageUrl:  vendor.imageUrl,
		Timestamp: vendor.createdAt,
	})
	return vendor, nil
}

func NewVendorFromHistory(events []event.DomainEvent) (*Vendor, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}
	
	vendor := &Vendor{}
	for _, ev := range events {
		if err := vendor.applyEvent(ev); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", ev.EventType(), err)
		}
	}
	return vendor, nil
}

// ReconstructVendor rebuilds a Vendor aggregate from database state WITHOUT raising events
func ReconstructVendor(id, name, email, phone, address, imageUrl string,
	version int, createdAt, updatedAt time.Time, isActive bool, bankAccount *VendorBankAccount) *Vendor {
	return &Vendor{
		id:                id,
		name:              name,
		email:             email,
		phone:             phone,
		address:           address,
		imageUrl:          imageUrl,
		bankAccount:       bankAccount,
		version:           version,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		isActive:          isActive,
		uncommittedEvents: nil, // No events when reconstructing from DB
	}
}

func (v *Vendor) UpdateProfile(name, email, phone, address string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	
	v.raiseEvent(&event.VendorUpdated{
		VendorID:     v.id,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Address:      address,
		EventVersion: v.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (v *Vendor) UpdateImageUrl(imageUrl string) error {
	if imageUrl == "" {
		return fmt.Errorf("imageUrl cannot be empty")
	}
	v.raiseEvent(&event.VendorImageUpdated{
		VendorID:     v.id,
		ImageUrl:     imageUrl,
		EventVersion: v.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

// UpdateBankAccount updates vendor's bank account information for payouts
func (v *Vendor) UpdateBankAccount(bankName, accountNumber, accountName, bankBranch string) error {
	if bankName == "" {
		return fmt.Errorf("bank name cannot be empty")
	}
	if accountNumber == "" {
		return fmt.Errorf("account number cannot be empty")
	}
	if accountName == "" {
		return fmt.Errorf("account name cannot be empty")
	}
	
	v.raiseEvent(&event.VendorBankAccountUpdated{
		VendorID:      v.id,
		BankName:      bankName,
		AccountNumber: accountNumber,
		AccountName:   accountName,
		BankBranch:    bankBranch,
		EventVersion:  v.version + 1,
		Timestamp:     time.Now(),
	})
	return nil
}

// HasBankAccount checks if vendor has bank account configured
func (v *Vendor) HasBankAccount() bool {
	return v.bankAccount != nil && v.bankAccount.AccountNumber != ""
}

// GetBankAccount returns vendor's bank account (safe copy)
func (v *Vendor) GetBankAccount() *VendorBankAccount {
	if v.bankAccount == nil {
		return nil
	}
	// Return copy to prevent external modification
	return &VendorBankAccount{
		BankName:      v.bankAccount.BankName,
		AccountNumber: v.bankAccount.AccountNumber,
		AccountName:   v.bankAccount.AccountName,
		BankBranch:    v.bankAccount.BankBranch,
	}
}

func (v *Vendor) Delete() error {
	v.raiseEvent(&event.VendorDeleted{
		VendorID:     v.id,
		EventVersion: v.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (v *Vendor) GetUncommittedEvents() []event.DomainEvent {
	return v.uncommittedEvents
}

func (v *Vendor) ClearUncommittedEvents() {
	v.uncommittedEvents = nil
}

func (v *Vendor) raiseEvent(ev event.DomainEvent) {
	v.uncommittedEvents = append(v.uncommittedEvents, ev)
	v.applyEvent(ev)
}

func (v *Vendor) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.VendorCreated:
		v.id = e.VendorID
		v.name = e.Name
		v.email = e.Email
		v.phone = e.Phone
		v.address = e.Address
		v.createdAt = e.Timestamp
		v.updatedAt = e.Timestamp
		v.version = 1
		v.isActive = true
		
	case *event.VendorUpdated:
		v.name = e.Name
		v.email = e.Email
		v.phone = e.Phone
		v.address = e.Address
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		
	case *event.VendorDeleted:
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		v.isActive = false
		
	case *event.VendorImageUpdated:
		v.imageUrl = e.ImageUrl
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		
	case *event.VendorBankAccountUpdated:
		v.bankAccount = &VendorBankAccount{
			BankName:      e.BankName,
			AccountNumber: e.AccountNumber,
			AccountName:   e.AccountName,
			BankBranch:    e.BankBranch,
		}
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
	
	return nil
}

// Getters
func (v *Vendor) ID() string           { return v.id }
func (v *Vendor) Name() string         { return v.name }
func (v *Vendor) Email() string        { return v.email }
func (v *Vendor) Phone() string        { return v.phone }
func (v *Vendor) Address() string      { return v.address }
func (v *Vendor) ImageUrl() string     { return v.imageUrl }
func (v *Vendor) Version() int         { return v.version }
func (v *Vendor) CreatedAt() time.Time { return v.createdAt }
func (v *Vendor) UpdatedAt() time.Time { return v.updatedAt }
func (v *Vendor) IsActive() bool       { return v.isActive }

// Entity interface implementation
func (v *Vendor) GetID() string    { return v.id }
func (v *Vendor) GetVersion() int  { return v.version }
func (v *Vendor) SetVersion(ver int) { v.version = ver }

// AggregateRoot interface implementation
func (v *Vendor) MarkEventsAsCommitted() {
	v.uncommittedEvents = nil
}

func (v *Vendor) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := v.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}
