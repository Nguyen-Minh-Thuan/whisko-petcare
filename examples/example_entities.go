package examples

import (
	"context"
	"time"
	"whisko-petcare/internal/infrastructure/mongo"
)

// Pet represents a pet entity for future use
type Pet struct {
	ID        string    `json:"id" bson:"_id"`
	Version   int       `json:"version" bson:"version"`
	OwnerID   string    `json:"owner_id" bson:"owner_id"`
	Name      string    `json:"name" bson:"name"`
	Species   string    `json:"species" bson:"species"`
	Breed     string    `json:"breed" bson:"breed"`
	Age       int       `json:"age" bson:"age"`
	Weight    float64   `json:"weight" bson:"weight"`
	Color     string    `json:"color" bson:"color"`
	Notes     string    `json:"notes" bson:"notes"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func (p *Pet) GetID() string          { return p.ID }
func (p *Pet) SetID(id string)        { p.ID = id }
func (p *Pet) GetVersion() int        { return p.Version }
func (p *Pet) SetVersion(version int) { p.Version = version }

// PetRepository extends the generic repository for Pet-specific queries
type PetRepository struct {
	*mongo.GenericRepository[*Pet]
}

func NewPetRepository() *PetRepository {
	return &PetRepository{
		GenericRepository: mongo.NewGenericRepository[*Pet](),
	}
}

func (r *PetRepository) FindByOwner(ctx context.Context, ownerID string) ([]*Pet, error) {
	return r.FindBy(ctx, func(pet *Pet) bool {
		return pet.OwnerID == ownerID
	})
}

func (r *PetRepository) FindBySpecies(ctx context.Context, species string) ([]*Pet, error) {
	return r.FindBy(ctx, func(pet *Pet) bool {
		return pet.Species == species
	})
}

// Service represents a grooming/spa service
type Service struct {
	ID          string    `json:"id" bson:"_id"`
	Version     int       `json:"version" bson:"version"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	Duration    int       `json:"duration" bson:"duration"` // in minutes
	Price       float64   `json:"price" bson:"price"`
	Category    string    `json:"category" bson:"category"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

func (s *Service) GetID() string          { return s.ID }
func (s *Service) SetID(id string)        { s.ID = id }
func (s *Service) GetVersion() int        { return s.Version }
func (s *Service) SetVersion(version int) { s.Version = version }

// ServiceRepository extends the generic repository for Service-specific queries
type ServiceRepository struct {
	*mongo.GenericRepository[*Service]
}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{
		GenericRepository: mongo.NewGenericRepository[*Service](),
	}
}

func (r *ServiceRepository) FindByCategory(ctx context.Context, category string) ([]*Service, error) {
	return r.FindBy(ctx, func(service *Service) bool {
		return service.Category == category
	})
}

func (r *ServiceRepository) FindByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*Service, error) {
	return r.FindBy(ctx, func(service *Service) bool {
		return service.Price >= minPrice && service.Price <= maxPrice
	})
}
