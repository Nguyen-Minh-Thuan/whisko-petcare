package event

import "time"

// ServiceCreated event
type ServiceCreated struct {
	ServiceID   string        `json:"service_id"`
	VendorID    string        `json:"vendor_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Price       int           `json:"price"`
	Duration    time.Duration `json:"duration"`
	Tags        []string      `json:"tags"`
	ImageUrl    string        `json:"image_url"`
	Timestamp   time.Time     `json:"timestamp"`
}

func (e *ServiceCreated) EventType() string     { return "ServiceCreated" }
func (e *ServiceCreated) AggregateID() string   { return e.ServiceID }
func (e *ServiceCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *ServiceCreated) Version() int          { return 1 }

// ServiceUpdated event
type ServiceUpdated struct {
	ServiceID    string        `json:"service_id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Price        int           `json:"price"`
	Duration     time.Duration `json:"duration"`
	Tags         []string      `json:"tags"`
	EventVersion int           `json:"version"`
	Timestamp    time.Time     `json:"timestamp"`
}

func (e *ServiceUpdated) EventType() string     { return "ServiceUpdated" }
func (e *ServiceUpdated) AggregateID() string   { return e.ServiceID }
func (e *ServiceUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *ServiceUpdated) Version() int          { return e.EventVersion }

// ServiceDeleted event
type ServiceDeleted struct {
	ServiceID    string    `json:"service_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *ServiceDeleted) EventType() string     { return "ServiceDeleted" }
func (e *ServiceDeleted) AggregateID() string   { return e.ServiceID }
func (e *ServiceDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *ServiceDeleted) Version() int          { return e.EventVersion }

// ServiceImageUpdated event
type ServiceImageUpdated struct {
	ServiceID    string    `json:"service_id"`
	ImageUrl     string    `json:"image_url"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *ServiceImageUpdated) EventType() string     { return "ServiceImageUpdated" }
func (e *ServiceImageUpdated) AggregateID() string   { return e.ServiceID }
func (e *ServiceImageUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *ServiceImageUpdated) Version() int          { return e.EventVersion }
