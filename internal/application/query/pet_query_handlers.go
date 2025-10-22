package query

import (
	"context"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/errors"
)

// GetPet represents a query to get a pet by ID
type GetPet struct {
	PetID string `json:"pet_id"`
}

// GetPetHandler handles get pet queries
type GetPetHandler struct {
	petProjection projection.PetProjection
}

// NewGetPetHandler creates a new get pet handler
func NewGetPetHandler(petProjection projection.PetProjection) *GetPetHandler {
	return &GetPetHandler{
		petProjection: petProjection,
	}
}

// Handle processes the get pet query
func (h *GetPetHandler) Handle(ctx context.Context, query *GetPet) (*projection.PetReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	if query.PetID == "" {
		return nil, errors.NewValidationError("pet_id is required")
	}

	pet, err := h.petProjection.GetByID(ctx, query.PetID)
	if err != nil {
		return nil, errors.NewNotFoundError("pet")
	}

	return pet, nil
}

// ListUserPets represents a query to list pets for a user
type ListUserPets struct {
	UserID string `json:"user_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// ListUserPetsHandler handles list user pets queries
type ListUserPetsHandler struct {
	petProjection projection.PetProjection
}

// NewListUserPetsHandler creates a new list user pets handler
func NewListUserPetsHandler(petProjection projection.PetProjection) *ListUserPetsHandler {
	return &ListUserPetsHandler{
		petProjection: petProjection,
	}
}

// Handle processes the list user pets query
func (h *ListUserPetsHandler) Handle(ctx context.Context, query *ListUserPets) ([]*projection.PetReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	if query.UserID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}

	// Set default pagination
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	pets, err := h.petProjection.GetByUserID(ctx, query.UserID, query.Offset, query.Limit)
	if err != nil {
		return nil, errors.NewInternalError("failed to list pets")
	}

	return pets, nil
}

// ListPets represents a query to list all pets
type ListPets struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// ListPetsHandler handles list pets queries
type ListPetsHandler struct {
	petProjection projection.PetProjection
}

// NewListPetsHandler creates a new list pets handler
func NewListPetsHandler(petProjection projection.PetProjection) *ListPetsHandler {
	return &ListPetsHandler{
		petProjection: petProjection,
	}
}

// Handle processes the list pets query
func (h *ListPetsHandler) Handle(ctx context.Context, query *ListPets) ([]*projection.PetReadModel, error) {
	if query == nil {
		return nil, errors.NewValidationError("query cannot be nil")
	}

	// Set default pagination
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	pets, err := h.petProjection.ListAll(ctx, query.Offset, query.Limit)
	if err != nil {
		return nil, errors.NewInternalError("failed to list pets")
	}

	return pets, nil
}
