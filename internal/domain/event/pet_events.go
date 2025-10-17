package event

import (
	"time"
)

type DomainEvent interface {
	EventType() string
	Timestamp() time.Time
	Version() int
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
	Timestamp time.Time `json:"timestamp"`
}