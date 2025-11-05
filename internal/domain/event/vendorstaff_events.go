package event

import "time"

// VendorStaffCreated event
type VendorStaffCreated struct {
	UserID    string    `json:"user_id"`
	VendorID  string    `json:"vendor_id"`
	Role      string    `json:"role"` // owner, manager, or staff
	Timestamp time.Time `json:"timestamp"`
}

func (e *VendorStaffCreated) EventType() string     { return "VendorStaffCreated" }
func (e *VendorStaffCreated) AggregateID() string   { return e.UserID + "-" + e.VendorID }
func (e *VendorStaffCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorStaffCreated) Version() int          { return 1 }

// VendorStaffRoleUpdated event
type VendorStaffRoleUpdated struct {
	UserID       string    `json:"user_id"`
	VendorID     string    `json:"vendor_id"`
	OldRole      string    `json:"old_role"`
	NewRole      string    `json:"new_role"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *VendorStaffRoleUpdated) EventType() string     { return "VendorStaffRoleUpdated" }
func (e *VendorStaffRoleUpdated) AggregateID() string   { return e.UserID + "-" + e.VendorID }
func (e *VendorStaffRoleUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorStaffRoleUpdated) Version() int          { return e.EventVersion }

// VendorStaffDeleted event
type VendorStaffDeleted struct {
	UserID       string    `json:"user_id"`
	VendorID     string    `json:"vendor_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *VendorStaffDeleted) EventType() string     { return "VendorStaffDeleted" }
func (e *VendorStaffDeleted) AggregateID() string   { return e.UserID + "-" + e.VendorID }
func (e *VendorStaffDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorStaffDeleted) Version() int          { return e.EventVersion }
