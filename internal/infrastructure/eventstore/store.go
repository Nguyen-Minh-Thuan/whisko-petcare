package eventstore

import (
	"context"
	"errors"
	"sync"
	"time"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
)

// MongoEventStore is a simple in-memory event store for demonstration.
// Replace with real MongoDB logic as needed.
type MongoEventStore struct {
	events map[string][]event.DomainEvent
	mutex  sync.RWMutex
}

// NewMongoEventStore returns a new in-memory event store.
func NewMongoEventStore() *MongoEventStore {
	return &MongoEventStore{
		events: make(map[string][]event.DomainEvent),
	}
}

// SaveEvents saves events for an aggregate.
func (s *MongoEventStore) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	currentEvents := s.events[aggregateID]
	if len(currentEvents) != expectedVersion {
		return errors.New("concurrency conflict: version mismatch")
	}
	s.events[aggregateID] = append(currentEvents, events...)
	return nil
}

// GetAllEvents returns all events for all aggregates.
func (s *MongoEventStore) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var allEvents []event.DomainEvent
	for _, evs := range s.events {
		allEvents = append(allEvents, evs...)
	}
	return allEvents, nil
}

// GetByID loads an aggregate.User by replaying its events.
func (s *MongoEventStore) GetByID(ctx context.Context, id string) (*aggregate.User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	events, ok := s.events[id]
	if !ok || len(events) == 0 {
		return nil, errors.New("user not found")
	}

	return aggregate.NewUserFromHistory(events)
}

// GetEvents returns all events for an aggregate.
func (s *MongoEventStore) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	evs, ok := s.events[aggregateID]
	if !ok {
		return nil, errors.New("not found")
	}
	return evs, nil
}

// Save saves all uncommitted events from the aggregate.User.
func (s *MongoEventStore) Save(ctx context.Context, user *aggregate.User) error {
	events := user.GetUncommittedEvents()
	if len(events) == 0 {
		return nil
	}
	expectedVersion := user.Version() - len(events)
	if err := s.SaveEvents(ctx, user.ID(), events, expectedVersion); err != nil {
		return err
	}
	user.ClearUncommittedEvents()
	return nil
}

// GetEventsSince returns events for an aggregate after a given version.
func (s *MongoEventStore) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	evs, ok := s.events[aggregateID]
	if !ok {
		return nil, errors.New("not found")
	}
	var result []event.DomainEvent
	for _, e := range evs {
		if e.Version() > version {
			result = append(result, e)
		}
	}
	return result, nil
}

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
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *UserDeleted) EventType() string     { return "UserDeleted" }
func (e *UserDeleted) AggregateID() string   { return e.UserID }
func (e *UserDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *UserDeleted) Version() int          { return 0 }
