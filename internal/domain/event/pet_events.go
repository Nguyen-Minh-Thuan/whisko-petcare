package event

import (
	"time"
)

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