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
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Address        string    `json:"address"`
	HashedPassword string    `json:"hashed_password"`
	Role           string    `json:"role"`
	ImageUrl       string    `json:"image_url"`
	IsActive       bool      `json:"is_active"`
	Timestamp      time.Time `json:"timestamp"`
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

// UserImageUpdated event
type UserImageUpdated struct {
	UserID       string    `json:"user_id"`
	ImageUrl     string    `json:"image_url"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserImageUpdated) EventType() string     { return "UserImageUpdated" }
func (e *UserImageUpdated) AggregateID() string   { return e.UserID }
func (e *UserImageUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserImageUpdated) Version() int          { return e.EventVersion }

// UserPasswordChanged event
type UserPasswordChanged struct {
	UserID         string    `json:"user_id"`
	HashedPassword string    `json:"hashed_password"`
	EventVersion   int       `json:"version"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e *UserPasswordChanged) EventType() string     { return "UserPasswordChanged" }
func (e *UserPasswordChanged) AggregateID() string   { return e.UserID }
func (e *UserPasswordChanged) OccurredAt() time.Time { return e.Timestamp }
func (e *UserPasswordChanged) Version() int          { return e.EventVersion }

// UserRoleUpdated event
type UserRoleUpdated struct {
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserRoleUpdated) EventType() string     { return "UserRoleUpdated" }
func (e *UserRoleUpdated) AggregateID() string   { return e.UserID }
func (e *UserRoleUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *UserRoleUpdated) Version() int          { return e.EventVersion }

// UserLoggedIn event
type UserLoggedIn struct {
	UserID       string    `json:"user_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *UserLoggedIn) EventType() string     { return "UserLoggedIn" }
func (e *UserLoggedIn) AggregateID() string   { return e.UserID }
func (e *UserLoggedIn) OccurredAt() time.Time { return e.Timestamp }
func (e *UserLoggedIn) Version() int          { return e.EventVersion }

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


