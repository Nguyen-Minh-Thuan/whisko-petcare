package event

import (
	"time"
)

// Health record types - defined here to avoid import cycles
type VaccinationRecord struct {
	ID           string    `json:"id"`
	VaccineName  string    `json:"vaccine_name"`
	Date         time.Time `json:"date"`
	NextDueDate  time.Time `json:"next_due_date,omitempty"`
	Veterinarian string    `json:"veterinarian,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

type MedicalRecord struct {
	ID           string    `json:"id"`
	Date         time.Time `json:"date"`
	Description  string    `json:"description"`
	Treatment    string    `json:"treatment,omitempty"`
	Veterinarian string    `json:"veterinarian,omitempty"`
	Diagnosis    string    `json:"diagnosis,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

type Allergy struct {
	ID            string    `json:"id"`
	Allergen      string    `json:"allergen"`
	Severity      string    `json:"severity"`
	Symptoms      string    `json:"symptoms,omitempty"`
	DiagnosedDate time.Time `json:"diagnosed_date,omitempty"`
	Notes         string    `json:"notes,omitempty"`
}

// PetCreated event
type PetCreated struct {
	PetID     string    `json:"pet_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Species   string    `json:"species"`
	Breed     string    `json:"breed"`
	Age       int       `json:"age"`
	Weight    float64   `json:"weight"`
	ImageUrl  string    `json:"image_url"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PetCreated) EventType() string     { return "PetCreated" }
func (e *PetCreated) AggregateID() string   { return e.PetID }
func (e *PetCreated) OccurredAt() time.Time { return e.Timestamp }
func (e *PetCreated) Version() int          { return 1 }

// PetUpdated event
type PetUpdated struct {
	PetID        string    `json:"pet_id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Species      string    `json:"species"`
	Breed        string    `json:"breed"`
	Age          int       `json:"age"`
	Weight       float64   `json:"weight"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PetUpdated) EventType() string     { return "PetUpdated" }
func (e *PetUpdated) AggregateID() string   { return e.PetID }
func (e *PetUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *PetUpdated) Version() int          { return e.EventVersion }

// PetDeleted event
type PetDeleted struct {
	PetID        string    `json:"pet_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PetDeleted) EventType() string     { return "PetDeleted" }
func (e *PetDeleted) AggregateID() string   { return e.PetID }
func (e *PetDeleted) OccurredAt() time.Time { return e.Timestamp }
func (e *PetDeleted) Version() int          { return e.EventVersion }

// PetImageUpdated event
type PetImageUpdated struct {
	PetID        string    `json:"pet_id"`
	ImageUrl     string    `json:"image_url"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PetImageUpdated) EventType() string     { return "PetImageUpdated" }
func (e *PetImageUpdated) AggregateID() string   { return e.PetID }
func (e *PetImageUpdated) OccurredAt() time.Time { return e.Timestamp }
func (e *PetImageUpdated) Version() int          { return e.EventVersion }

// PetVaccinationAdded event
type PetVaccinationAdded struct {
	PetID        string            `json:"pet_id"`
	Record       VaccinationRecord `json:"record"`
	EventVersion int               `json:"version"`
	Timestamp    time.Time         `json:"timestamp"`
}

func (e *PetVaccinationAdded) EventType() string     { return "PetVaccinationAdded" }
func (e *PetVaccinationAdded) AggregateID() string   { return e.PetID }
func (e *PetVaccinationAdded) OccurredAt() time.Time { return e.Timestamp }
func (e *PetVaccinationAdded) Version() int          { return e.EventVersion }

// PetMedicalRecordAdded event
type PetMedicalRecordAdded struct {
	PetID        string        `json:"pet_id"`
	Record       MedicalRecord `json:"record"`
	EventVersion int           `json:"version"`
	Timestamp    time.Time     `json:"timestamp"`
}

func (e *PetMedicalRecordAdded) EventType() string     { return "PetMedicalRecordAdded" }
func (e *PetMedicalRecordAdded) AggregateID() string   { return e.PetID }
func (e *PetMedicalRecordAdded) OccurredAt() time.Time { return e.Timestamp }
func (e *PetMedicalRecordAdded) Version() int          { return e.EventVersion }

// PetAllergyAdded event
type PetAllergyAdded struct {
	PetID        string    `json:"pet_id"`
	Allergy      Allergy   `json:"allergy"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PetAllergyAdded) EventType() string     { return "PetAllergyAdded" }
func (e *PetAllergyAdded) AggregateID() string   { return e.PetID }
func (e *PetAllergyAdded) OccurredAt() time.Time { return e.Timestamp }
func (e *PetAllergyAdded) Version() int          { return e.EventVersion }

// PetAllergyRemoved event
type PetAllergyRemoved struct {
	PetID        string    `json:"pet_id"`
	AllergyID    string    `json:"allergy_id"`
	EventVersion int       `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *PetAllergyRemoved) EventType() string     { return "PetAllergyRemoved" }
func (e *PetAllergyRemoved) AggregateID() string   { return e.PetID }
func (e *PetAllergyRemoved) OccurredAt() time.Time { return e.Timestamp }
func (e *PetAllergyRemoved) Version() int          { return e.EventVersion }