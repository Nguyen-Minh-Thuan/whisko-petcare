package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"
)

type VendorStaff struct {
	userID    string
	vendorID  string
	version   int
	createdAt time.Time
	updatedAt time.Time
	isActive  bool

	uncommittedEvents []event.DomainEvent
}

func NewVendorStaff(userID, vendorID string) (*VendorStaff, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if vendorID == "" {
		return nil, fmt.Errorf("vendorID cannot be empty")
	}
	vendorStaff := &VendorStaff{
		userID:    userID,
		vendorID:  vendorID,
		version:   1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}	
	vendorStaff.raiseEvent(&event.VendorStaffCreated{
		UserID:    userID,
		VendorID:  vendorID,
		Timestamp: vendorStaff.createdAt,
	})
	return vendorStaff, nil
}

func NewVendorStaffFromHistory(events []event.DomainEvent) (*VendorStaff, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}
	
	vendorStaff := &VendorStaff{}
	for _, ev := range events {
		if err := vendorStaff.applyEvent(ev); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", ev.EventType(), err)
		}
	}
	return vendorStaff, nil
}

func (v *VendorStaff) Delete() error {
	v.raiseEvent(&event.VendorStaffDeleted{
		UserID:       v.userID,
		VendorID:     v.vendorID,
		EventVersion: v.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (v *VendorStaff) GetUncommittedEvents() []event.DomainEvent {
	return v.uncommittedEvents
}

func (v *VendorStaff) ClearUncommittedEvents() {
	v.uncommittedEvents = nil
}

func (v *VendorStaff) raiseEvent(ev event.DomainEvent) {
	v.uncommittedEvents = append(v.uncommittedEvents, ev)
	v.applyEvent(ev)
}

func (v *VendorStaff) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.VendorStaffCreated:
		v.userID = e.UserID
		v.vendorID = e.VendorID
		v.createdAt = e.Timestamp
		v.updatedAt = e.Timestamp
		v.version = 1
		v.isActive = true
		
	case *event.VendorStaffDeleted:
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		v.isActive = false
		
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}
	
	return nil
}

// Getters
func (v *VendorStaff) UserID() string       { return v.userID }
func (v *VendorStaff) VendorID() string     { return v.vendorID }
func (v *VendorStaff) Version() int         { return v.version }
func (v *VendorStaff) CreatedAt() time.Time { return v.createdAt }
func (v *VendorStaff) UpdatedAt() time.Time { return v.updatedAt }
func (v *VendorStaff) IsActive() bool       { return v.isActive }

// Entity interface implementation
func (v *VendorStaff) GetID() string    { return v.userID + "-" + v.vendorID }
func (v *VendorStaff) GetVersion() int  { return v.version }
func (v *VendorStaff) SetVersion(ver int) { v.version = ver }

// AggregateRoot interface implementation
func (v *VendorStaff) MarkEventsAsCommitted() {
	v.uncommittedEvents = nil
}

func (v *VendorStaff) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := v.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}

