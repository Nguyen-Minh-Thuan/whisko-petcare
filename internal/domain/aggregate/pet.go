package aggregate

import (
	"fmt"
	"time"
	"whisko-petcare/internal/domain/event"

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
	imageUrl         string
	version          int
	createdAt        time.Time
	updatedAt        time.Time
	isActive         bool

	// Health data
	vaccinationRecords []event.VaccinationRecord
	medicalHistory     []event.MedicalRecord
	allergies          []event.Allergy

	uncommittedEvents []event.DomainEvent
}

func NewPet(userID, name, species, breed string, age int, weight float64, imageUrl ...string) (*Pet, error) {
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

	// Set imageUrl if provided
	if len(imageUrl) > 0 && imageUrl[0] != "" {
		pet.imageUrl = imageUrl[0]
	}

	pet.raiseEvent(&event.PetCreated{
		PetID:     pet.id,
		UserID:    userID,
		Name:      name,
		Species:   species,
		Breed:     breed,
		Age:       age,
		Weight:    weight,
		ImageUrl:  pet.imageUrl,
		Timestamp: pet.createdAt,
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
		PetID:        p.id,
		UserID:       p.userID,
		Name:         name,
		Species:      species,
		Breed:        breed,
		Age:          age,
		Weight:       weight,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) UpdateImageUrl(imageUrl string) error {
	if imageUrl == "" {
		return fmt.Errorf("imageUrl cannot be empty")
	}
	p.raiseEvent(&event.PetImageUpdated{
		PetID:        p.id,
		ImageUrl:     imageUrl,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) Delete() error {
	p.raiseEvent(&event.PetDeleted{
		PetID:        p.id,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

// Health management methods

func (p *Pet) AddVaccinationRecord(vaccineName string, date, nextDueDate time.Time, veterinarian, notes string) error {
	if vaccineName == "" {
		return fmt.Errorf("vaccine name cannot be empty")
	}
	if date.IsZero() {
		return fmt.Errorf("vaccination date cannot be empty")
	}

	record := event.VaccinationRecord{
		ID:           uuid.New().String(),
		VaccineName:  vaccineName,
		Date:         date,
		NextDueDate:  nextDueDate,
		Veterinarian: veterinarian,
		Notes:        notes,
	}

	p.raiseEvent(&event.PetVaccinationAdded{
		PetID:        p.id,
		Record:       record,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) AddMedicalRecord(date time.Time, description, treatment, veterinarian, diagnosis, notes string) error {
	if date.IsZero() {
		return fmt.Errorf("medical record date cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	record := event.MedicalRecord{
		ID:           uuid.New().String(),
		Date:         date,
		Description:  description,
		Treatment:    treatment,
		Veterinarian: veterinarian,
		Diagnosis:    diagnosis,
		Notes:        notes,
	}

	p.raiseEvent(&event.PetMedicalRecordAdded{
		PetID:        p.id,
		Record:       record,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) AddAllergy(allergen, severity, symptoms string, diagnosedDate time.Time, notes string) error {
	if allergen == "" {
		return fmt.Errorf("allergen cannot be empty")
	}
	if severity == "" {
		severity = "mild"
	}
	// Validate severity
	if severity != "mild" && severity != "moderate" && severity != "severe" {
		return fmt.Errorf("invalid severity: must be 'mild', 'moderate', or 'severe'")
	}

	allergy := event.Allergy{
		ID:            uuid.New().String(),
		Allergen:      allergen,
		Severity:      severity,
		Symptoms:      symptoms,
		DiagnosedDate: diagnosedDate,
		Notes:         notes,
	}

	p.raiseEvent(&event.PetAllergyAdded{
		PetID:        p.id,
		Allergy:      allergy,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) RemoveAllergy(allergyID string) error {
	if allergyID == "" {
		return fmt.Errorf("allergy ID cannot be empty")
	}

	// Check if allergy exists
	found := false
	for _, a := range p.allergies {
		if a.ID == allergyID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("allergy not found: %s", allergyID)
	}

	p.raiseEvent(&event.PetAllergyRemoved{
		PetID:        p.id,
		AllergyID:    allergyID,
		EventVersion: p.version + 1,
		Timestamp:    time.Now(),
	})
	return nil
}

func (p *Pet) GetUncommittedEvents() []event.DomainEvent {
	return p.uncommittedEvents
}

func (p *Pet) ClearUncommittedEvents() {
	p.uncommittedEvents = nil
}

func (p *Pet) raiseEvent(ev event.DomainEvent) {
	p.uncommittedEvents = append(p.uncommittedEvents, ev)
	_ = p.applyEvent(ev)
}

// applyEvent applies an event to the pet state
func (p *Pet) applyEvent(ev event.DomainEvent) error {
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
		p.version = 1
		p.isActive = true
		
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
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
		
	case *event.PetDeleted:
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
		p.isActive = false
		
	case *event.PetImageUpdated:
		p.imageUrl = e.ImageUrl
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
	
	case *event.PetVaccinationAdded:
		p.vaccinationRecords = append(p.vaccinationRecords, e.Record)
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
	
	case *event.PetMedicalRecordAdded:
		p.medicalHistory = append(p.medicalHistory, e.Record)
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
	
	case *event.PetAllergyAdded:
		p.allergies = append(p.allergies, e.Allergy)
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
	
	case *event.PetAllergyRemoved:
		// Remove allergy from slice
		for i, a := range p.allergies {
			if a.ID == e.AllergyID {
				p.allergies = append(p.allergies[:i], p.allergies[i+1:]...)
				break
			}
		}
		p.version = e.EventVersion
		p.updatedAt = e.Timestamp
		
	default:
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
func (p *Pet) ImageUrl() string  { return p.imageUrl }
func (p *Pet) Version() int      { return p.version }
func (p *Pet) CreatedAt() time.Time { return p.createdAt }
func (p *Pet) UpdatedAt() time.Time { return p.updatedAt }
func (p *Pet) IsActive() bool       { return p.isActive }

// Health data getters
func (p *Pet) VaccinationRecords() []event.VaccinationRecord { return p.vaccinationRecords }
func (p *Pet) MedicalHistory() []event.MedicalRecord         { return p.medicalHistory }
func (p *Pet) Allergies() []event.Allergy                    { return p.allergies }

//Entity interface implementation
func (p *Pet) GetID() string    { return p.id }
func (p *Pet) GetVersion() int  { return p.version }
func (p *Pet) SetVersion(v int) { p.version = v }
func (p *Pet) MarkInactive() 	{ p.isActive = false }

// Repository helper methods - for database reconstruction only
func (p *Pet) SetVaccinationRecords(records []event.VaccinationRecord) { p.vaccinationRecords = records }
func (p *Pet) SetMedicalHistory(records []event.MedicalRecord)         { p.medicalHistory = records }
func (p *Pet) SetAllergies(allergies []event.Allergy)                  { p.allergies = allergies }

func (p *Pet) MarkEventsAsCommitted(){
	p.uncommittedEvents = nil
}

func (p *Pet) LoadFromHistory(events []event.DomainEvent) error {
	for _, e := range events {
		if err := p.applyEvent(e); err != nil {
			return fmt.Errorf("failed to apply event: %w", err)
		}
	}
	return nil
}	
