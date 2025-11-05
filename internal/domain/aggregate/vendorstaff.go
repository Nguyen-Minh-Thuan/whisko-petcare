package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"
)

// VendorStaffRole represents the role of a vendor staff member
type VendorStaffRole string

const (
	VendorStaffRoleOwner   VendorStaffRole = "owner"   // First creator, full control
	VendorStaffRoleManager VendorStaffRole = "manager" // Can manage staff, assigned by owner
	VendorStaffRoleStaff   VendorStaffRole = "staff"   // Default role for added staff
)

type VendorStaff struct {
	userID    string
	vendorID  string
	role      VendorStaffRole
	version   int
	createdAt time.Time
	updatedAt time.Time
	isActive  bool

	uncommittedEvents []event.DomainEvent
}

// NewVendorStaff creates a new vendor staff with the specified role
func NewVendorStaff(userID, vendorID string, role VendorStaffRole) (*VendorStaff, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if vendorID == "" {
		return nil, fmt.Errorf("vendorID cannot be empty")
	}
	if role == "" {
		role = VendorStaffRoleStaff // Default role
	}
	vendorStaff := &VendorStaff{
		userID:    userID,
		vendorID:  vendorID,
		role:      role,
		version:   1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}	
	vendorStaff.raiseEvent(&event.VendorStaffCreated{
		UserID:    userID,
		VendorID:  vendorID,
		Role:      string(role),
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

// UpdateRole updates the role of the vendor staff member
func (v *VendorStaff) UpdateRole(newRole VendorStaffRole) error {
	if newRole == "" {
		return fmt.Errorf("role cannot be empty")
	}
	if v.role == newRole {
		return nil // No change needed
	}
	
	v.raiseEvent(&event.VendorStaffRoleUpdated{
		UserID:       v.userID,
		VendorID:     v.vendorID,
		OldRole:      string(v.role),
		NewRole:      string(newRole),
		EventVersion: v.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
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
		v.role = VendorStaffRole(e.Role)
		v.createdAt = e.Timestamp
		v.updatedAt = e.Timestamp
		v.version = 1
		v.isActive = true
		
	case *event.VendorStaffRoleUpdated:
		v.role = VendorStaffRole(e.NewRole)
		v.version = e.EventVersion
		v.updatedAt = e.Timestamp
		
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
func (v *VendorStaff) UserID() string           { return v.userID }
func (v *VendorStaff) VendorID() string         { return v.vendorID }
func (v *VendorStaff) Role() VendorStaffRole    { return v.role }
func (v *VendorStaff) Version() int             { return v.version }
func (v *VendorStaff) CreatedAt() time.Time     { return v.createdAt }
func (v *VendorStaff) UpdatedAt() time.Time     { return v.updatedAt }
func (v *VendorStaff) IsActive() bool           { return v.isActive }

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

