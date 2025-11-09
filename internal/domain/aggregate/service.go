package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

	"github.com/google/uuid"
)

type Service struct {
	id          string
	vendorId    string
	name        string
	description string
	price       int // Price in VND cents
	duration    time.Duration
	imageUrl    string
	tags        []string
	createdAt   time.Time
	updatedAt   time.Time
	version     int
	isActive    bool
	
	uncommittedEvents []event.DomainEvent
}

func NewService(vendorId, name, description string, price int, duration time.Duration, tags []string, imageUrl ...string) (*Service, error) {
	// Validate input
	if vendorId == "" {
		return nil, fmt.Errorf("vendorId cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if duration <= 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}
	if tags == nil {
		tags = []string{}
	}
	service := &Service{
		id:          uuid.New().String(),
		vendorId:    vendorId,
		name:        name,
		description: description,
		price:       price,
		duration:    duration,
		tags:        tags,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),	
		version:     1,
		isActive:    true,
	}
	
	// Set imageUrl if provided
	if len(imageUrl) > 0 && imageUrl[0] != "" {
		service.imageUrl = imageUrl[0]
	}
	
	service.raiseEvent(&event.ServiceCreated{
		ServiceID:   service.id,
		VendorID:    service.vendorId,
		Name:        name,
		Description: description,
		Price:       price,
		Duration:    duration,
		Tags:        tags,
		ImageUrl:    service.imageUrl,
		Timestamp:   service.createdAt,
	})
	return service, nil
}

func NewServiceFromHistory(events []event.DomainEvent) (*Service, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}
	
	service := &Service{}
	for _, e := range events {
		if err := service.applyEvent(e); err != nil {
			return nil, fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return service, nil
}

// ReconstructService rebuilds a Service aggregate from database state WITHOUT raising events
func ReconstructService(id, vendorID, name, description, imageUrl string, price int, duration time.Duration, tags []string,
	version int, createdAt, updatedAt time.Time, isActive bool) *Service {
	return &Service{
		id:                id,
		vendorId:          vendorID,
		name:              name,
		description:       description,
		price:             price,
		duration:          duration,
		tags:              tags,
		imageUrl:          imageUrl,
		version:           version,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		isActive:          isActive,
		uncommittedEvents: nil, // No events when reconstructing from DB
	}
}

func (s *Service) UpdateService(name, description string, price int, duration time.Duration, tags []string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	if duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if tags == nil {
		tags = []string{}
	}

	s.raiseEvent(&event.ServiceUpdated{
		ServiceID:    s.id,
		Name:         name,
		Description:  description,
		Price:        price,
		Duration:     duration,
		Tags:         tags,
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (s *Service) UpdateImageUrl(imageUrl string) error {
	if imageUrl == "" {
		return fmt.Errorf("imageUrl cannot be empty")
	}
	s.raiseEvent(&event.ServiceImageUpdated{
		ServiceID:    s.id,
		ImageUrl:     imageUrl,
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (s *Service) Delete() error {
	s.raiseEvent(&event.ServiceDeleted{
		ServiceID:    s.id,
		EventVersion: s.version + 1,
		Timestamp:    time.Now(),
	})
	
	return nil
}

func (s *Service) GetUncommittedEvents() []event.DomainEvent {
	return s.uncommittedEvents
}

func (s *Service) ClearUncommittedEvents() {
	s.uncommittedEvents = nil
}

func (s *Service) raiseEvent(ev event.DomainEvent) {
	s.uncommittedEvents = append(s.uncommittedEvents, ev)
	s.applyEvent(ev)
}

func (s *Service) applyEvent(ev event.DomainEvent) error {
	switch e := ev.(type) {
	case *event.ServiceCreated:
		s.id = e.ServiceID
		s.vendorId = e.VendorID
		s.name = e.Name
		s.description = e.Description
		s.price = e.Price
		s.duration = e.Duration
		s.tags = e.Tags
		s.createdAt = e.Timestamp
		s.updatedAt = e.Timestamp
		s.version = 1
		s.isActive = true
		
	case *event.ServiceUpdated:
		s.name = e.Name
		s.description = e.Description
		s.price = e.Price
		s.duration = e.Duration
		s.tags = e.Tags
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		
	case *event.ServiceDeleted:
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		s.isActive = false
		
	case *event.ServiceImageUpdated:
		s.imageUrl = e.ImageUrl
		s.version = e.EventVersion
		s.updatedAt = e.Timestamp
		
	default:
		return fmt.Errorf("unknown event type: %T", ev)
	}

	return nil
}

// Getters
func (s *Service) ID() string             { return s.id }
func (s *Service) VendorID() string       { return s.vendorId }
func (s *Service) Name() string           { return s.name }
func (s *Service) Description() string    { return s.description }
func (s *Service) Price() int             { return s.price }
func (s *Service) Duration() time.Duration { return s.duration }
func (s *Service) ImageUrl() string       { return s.imageUrl }
func (s *Service) Tags() []string         { return s.tags }
func (s *Service) CreatedAt() time.Time   { return s.createdAt }
func (s *Service) UpdatedAt() time.Time   { return s.updatedAt }
func (s *Service) Version() int           { return s.version }
func (s *Service) IsActive() bool         { return s.isActive }

// Entity interface implementation
func (s *Service) GetID() string    { return s.id }
func (s *Service) GetVersion() int  { return s.version }
func (s *Service) SetVersion(v int) { s.version = v }

// AggregateRoot interface implementation
func (s *Service) MarkEventsAsCommitted() {
	s.uncommittedEvents = nil
}

func (s *Service) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := s.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", e.EventType(), err)
		}
	}
	return nil
}
