package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"
	"github.com/google/uuid"
)

type Vendor struct {
	id        string
	name      string
	email     string
	phone     string
	address   string
	version   int
	createdAt time.Time
	updatedAt time.Time
	isActive  bool

	uncommittedEvents []event.DomainEvent
}

func NewVendor(name, email, phone, address string) (*Vendor, error) {
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
		id:        uuid.New().String(),
		name:      name,
		email:     email,
		phone:     phone,
		address:   address,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		version:   1,
		isActive:  true,
	}

	vendor.raiseEvent(&event.VendorCreated{
		VendorID:  vendor.id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		Address:   address,
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
