package aggregate

import(
	"time"
	"whisko-petcare/internal/domain/event"
	"fmt"

	"github.com/google/uuid"
)

type Pet struct {
	id               string
	userID           string
	name             string
	species          string
	breed            string
	description      string
	age              int
	weight           float64
	version          int
	createdAt        time.Time
	updatedAt        time.Time
	isActive         bool

	uncommittedEvents []event.DomainEvent
}

func NewPet(userID, name, species, breed string, age int, weight float64) (*Pet, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if age < 0 {
		return nil, fmt.Errorf("invalid age: %d", age)
	}
	if weight < 0 {
		return nil, fmt.Errorf("invalid weight: %f", weight)
	}

	pet := &Pet{
		id:        uuid.New().String(),
		userID:    userID,
		name:      name,
		species:   species,
		breed:     breed,
		age:       age,
		weight:    weight,
		version:   1,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}

	pet.raiseEvent(&event.PetCreated{
		PetID:     pet.id,
		UserID:    userID,
		Name:      name,
		Species:   species,
		Breed:     breed,
		Age:       age,
		Weight:    weight,
	})

	return pet, nil
}

func NewPetFromHistory(events []event.DomainEvent) (*Pet, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}
	pet := &Pet{}
	for _, e := range events {
		if err := pet.applyEvent(e); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
		pet.version = e.Version()
	}

	return pet, nil
}

func (p *Pet) UpdateProfile(name, species, breed string, age int, weight float64) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if age < 0 {
		return fmt.Errorf("invalid age: %d", age)
	}
	if weight < 0 {
		return fmt.Errorf("invalid weight: %f", weight)
	}
	p.raiseEvent(&event.PetUpdated{
		PetID:     p.id,
		Name:      name,
		Species:   species,
		Breed:     breed,
		Age:       age,
		Weight:    weight,
		Version:   p.version + 1,
		Timestamp: time.Now(),
	})
	return nil
}

func (p *Pet) Delete() error {
	p.raiseEvent(&event.PetDeleted{
		PetID:     p.id,
		Version:   p.version + 1,
		Timestamp: time.Now(),
		IsActive:  false,
	})
	return nil
}

func (p *pet) GetUncommittedEvents() []event.DomainEvent {
	return p.uncommittedEvents
}

func (p *Pet) ClearUncommittedEvents() {
	p.uncommittedEvents = nil
}

func (p *Pet) raiseEvent(ev event.DomainEvent) {
	p.uncommittedEvents = append(p.uncommittedEvents, ev)
	p.applyEvent(ev)
}

// applyEvent applies an event to the pet state
func (p *Pet) applyEvent(ev event.DomainEvent) {
	p.uncommittedEvents = append(p.uncommittedEvents, ev)
	switch e := ev.(type) {
	case *event.PetCreated:
		p.id = e.PetID
		p.userID = e.UserID
		p.name = e.Name
		p.species = e.Species
		p.breed = e.Breed
		p.age = e.Age
		p.weight = e.Weight
		p.createdAt = e.Timestamp
		p.updatedAt = e.Timestamp
	case *event.PetUpdated:
		if e.Name != "" {
			p.name = e.Name
		}
		if e.Species != "" {
			p.species = e.Species
		}
		if e.Breed != "" {
			p.breed = e.Breed
		}
		if e.Age != 0 {
			p.age = e.Age
		}
		if e.Weight != 0 {
			p.weight = e.Weight
		}
		p.updatedAt = e.Timestamp
	case *event.PetDeleted:
		p.updatedAt = e.Timestamp
		p.version = e.Version
		p.isActive = e.IsActive
		// Handle deletion logic if needed
	default:
		// Unknown event type
		return fmt.Errorf("unknown event type: %T", ev)
	}

	return nil

}

//Getters
func (p *Pet) ID() string        { return p.id }
func (p *Pet) UserID() string    { return p.userID }
func (p *Pet) Name() string      { return p.name }
func (p *Pet) Species() string   { return p.species }
func (p *Pet) Breed() string     { return p.breed }
func (p *Pet) Age() int          { return p.age }
func (p *Pet) Weight() float64   { return p.weight }
func (p *Pet) Version() int      { return p.version }
func (p *Pet) CreatedAt() time.Time { return p.createdAt }
func (p *Pet) UpdatedAt() time.Time { return p.updatedAt }
func (p *Pet) IsActive() bool       { return p.isActive }

//Entity interface implementation
func (p *Pet) GetID() string    { return p.id }
func (p *Pet) GetVersion() int  { return p.version }
func (p *Pet) SetVersion(v int) { p.version = v }
func (p *Pet) MarkInactive() 	{ p.isActive = false }

func (p *Pet) MarkEventsAsCommitted(){
	p.uncommittedEvents = nil
}

func (p *Pet) raiseEvent(events []event.DomainEvent) error{
	for _, ev := range events {
		if err := p.applyEvent(ev); err != nil {
			return fmt.Errorf("failed to apply event %s: %w", ev.EventType(), err)
		}
	}
	return nil
}

