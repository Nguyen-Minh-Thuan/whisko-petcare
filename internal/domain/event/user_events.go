package event

import "time"

// DomainEvent represents a domain event
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	Version() int
}

// UserCreated event
type UserCreated struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *UserCreated) EventType() string     { return "UserCreated" }
func (e *UserCreated) AggregateID() string   { return e.UserID }
func (e *UserCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserCreated) Version() int          { return 1 }

// UserProfileUpdated event
type UserProfileUpdated struct {
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserProfileUpdated) EventType() string     { return "UserProfileUpdated" }
func (e *UserProfileUpdated) AggregateID() string   { return e.UserID }
func (e *UserProfileUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserProfileUpdated) Version() int          { return e.EventVersion }

// UserContactUpdated event
type UserContactUpdated struct {
	UserID       string    `json:"user_id"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserContactUpdated) EventType() string     { return "UserContactUpdated" }
func (e *UserContactUpdated) AggregateID() string   { return e.UserID }
func (e *UserContactUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserContactUpdated) Version() int          { return e.EventVersion }

// UserDeleted event
type UserDeleted struct {
	UserID       string    `json:"user_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserDeleted) EventType() string     { return "UserDeleted" }
func (e *UserDeleted) AggregateID() string   { return e.UserID }
func (e *UserDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *UserDeleted) Version() int          { return e.EventVersion }


