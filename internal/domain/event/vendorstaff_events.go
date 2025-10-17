package event

import "time"

// VendorStaffCreated event
type VendorStaffCreated struct {
	UserID    string    `json:"user_id"`
	VendorID  string    `json:"vendor_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *VendorStaffCreated) EventType() string     { return "VendorStaffCreated" }
func (e *VendorStaffCreated) AggregateID() string   { return e.UserID + "-" + e.VendorID }
func (e *VendorStaffCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorStaffCreated) Version() int          { return 1 }

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
