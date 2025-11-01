package event

import "time"

// VendorCreated event
type VendorCreated struct {
	VendorID  string    `json:"vendor_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	ImageUrl  string    `json:"image_url"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *VendorCreated) EventType() string     { return "VendorCreated" }
func (e *VendorCreated) AggregateID() string   { return e.VendorID }
func (e *VendorCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorCreated) Version() int          { return 1 }

// VendorUpdated event
type VendorUpdated struct {
	VendorID     string    `json:"vendor_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *VendorUpdated) EventType() string     { return "VendorUpdated" }
func (e *VendorUpdated) AggregateID() string   { return e.VendorID }
func (e *VendorUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorUpdated) Version() int          { return e.EventVersion }

// VendorDeleted event
type VendorDeleted struct {
	VendorID     string    `json:"vendor_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *VendorDeleted) EventType() string     { return "VendorDeleted" }
func (e *VendorDeleted) AggregateID() string   { return e.VendorID }
func (e *VendorDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorDeleted) Version() int          { return e.EventVersion }

// VendorImageUpdated event
type VendorImageUpdated struct {
	VendorID     string    `json:"vendor_id"`
	ImageUrl     string    `json:"image_url"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *VendorImageUpdated) EventType() string     { return "VendorImageUpdated" }
func (e *VendorImageUpdated) AggregateID() string   { return e.VendorID }
func (e *VendorImageUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *VendorImageUpdated) Version() int          { return e.EventVersion }
