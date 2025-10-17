package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

	"github.com/google/uuid"
)

type User struct {
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

func NewUser(id, name, email string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user := &User{
		id:        uuid.New().String(),
		name:      name,
		email:     email,
		version:   1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}

	user.raiseEvent(&event.UserCreated{
		UserID:    id,
		Name:      name,
		Email:     email,
		Phone:     user.phone,
		Address:   user.address,
		Timestamp: user.createdAt,
	})

	return user, nil
}

func NewUserFromHistory(events []event.DomainEvent) (*User, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	user := &User{}

	for _, e := range events {
		if err := user.applyEvent(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}

	return user, nil
}

func (u *User) UpdateProfile(name, email string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	u.raiseEvent(&event.UserProfileUpdated{
		UserID:       u.id,
		Name:         name,
		Email:        email,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) UpdateContactInfo(phone, address string) error {
	u.raiseEvent(&event.UserContactUpdated{
		UserID:       u.id,
		Phone:        phone,
		Address:      address,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) Delete() error {
	u.raiseEvent(&event.UserDeleted{
		UserID:       u.id,
		EventVersion: u.version + 1,
		Timestamp:    time.Now(),
	})

	return nil
}

func (u *User) GetUncommittedEvents() []event.DomainEvent {
	return u.uncommittedEvents
}

func (u *User) ClearUncommittedEvents() {
	u.uncommittedEvents = nil
}

func (u *User) raiseEvent(ev event.DomainEvent) {
	u.uncommittedEvents = append(u.uncommittedEvents, ev)
	u.applyEvent(ev)
}

func (u *User) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.UserCreated:
		u.id = e.UserID
		u.name = e.Name
		u.email = e.Email
		u.phone = e.Phone
		u.address = e.Address
		u.version = 1
		u.createdAt = e.Timestamp
		u.updatedAt = e.Timestamp

	case *event.UserProfileUpdated:
		u.name = e.Name
		u.email = e.Email
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp

	case *event.UserContactUpdated:
		u.phone = e.Phone
		u.address = e.Address
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp

	case *event.UserDeleted:
		u.version = e.EventVersion
		u.updatedAt = e.Timestamp
		u.isActive = false

	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}

	return nil
}

// Getters
func (u *User) ID() string           { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) Phone() string        { return u.phone }
func (u *User) Address() string      { return u.address }
func (u *User) Version() int         { return u.version }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Entity interface implementation
func (u *User) GetID() string    { return u.id }
func (u *User) SetID(id string)  { u.id = id }
func (u *User) GetVersion() int  { return u.version }
func (u *User) SetVersion(v int) { u.version = v }	

// AggregateRoot interface implementation
func (u *User) MarkEventsAsCommitted() {
	u.uncommittedEvents = nil
}

func (u *User) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := u.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
	}
