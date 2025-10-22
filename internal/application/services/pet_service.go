package services

import (
	"context"
	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/infrastructure/projection"
)

// PetService orchestrates pet operations
type PetService struct {
	// Command handlers (using Unit of Work)
	createPetHandler *command.CreatePetWithUoWHandler
	updatePetHandler *command.UpdatePetWithUoWHandler
	deletePetHandler *command.DeletePetWithUoWHandler

	// Query handlers (using Projections)
	getPetHandler       *query.GetPetHandler
	listUserPetsHandler *query.ListUserPetsHandler
	listPetsHandler     *query.ListPetsHandler
}

// NewPetService creates a new pet service
func NewPetService(
	createPetHandler *command.CreatePetWithUoWHandler,
	updatePetHandler *command.UpdatePetWithUoWHandler,
	deletePetHandler *command.DeletePetWithUoWHandler,
	getPetHandler *query.GetPetHandler,
	listUserPetsHandler *query.ListUserPetsHandler,
	listPetsHandler *query.ListPetsHandler,
) *PetService {
	return &PetService{
		createPetHandler:    createPetHandler,
		updatePetHandler:    updatePetHandler,
		deletePetHandler:    deletePetHandler,
		getPetHandler:       getPetHandler,
		listUserPetsHandler: listUserPetsHandler,
		listPetsHandler:     listPetsHandler,
	}
}

// Command operations

// CreatePet creates a new pet
func (s *PetService) CreatePet(ctx context.Context, cmd command.CreatePet) error {
	return s.createPetHandler.Handle(ctx, &cmd)
}

// UpdatePet updates pet information
func (s *PetService) UpdatePet(ctx context.Context, cmd command.UpdatePet) error {
	return s.updatePetHandler.Handle(ctx, &cmd)
}

// DeletePet deletes a pet
func (s *PetService) DeletePet(ctx context.Context, cmd command.DeletePet) error {
	return s.deletePetHandler.Handle(ctx, &cmd)
}

// Query operations

// GetPet retrieves a pet by ID
func (s *PetService) GetPet(ctx context.Context, petID string) (*projection.PetReadModel, error) {
	return s.getPetHandler.Handle(ctx, &query.GetPet{PetID: petID})
}

// ListUserPets retrieves all pets for a user
func (s *PetService) ListUserPets(ctx context.Context, userID string, offset, limit int) ([]*projection.PetReadModel, error) {
	return s.listUserPetsHandler.Handle(ctx, &query.ListUserPets{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})
}

// ListPets retrieves all pets
func (s *PetService) ListPets(ctx context.Context, offset, limit int) ([]*projection.PetReadModel, error) {
	return s.listPetsHandler.Handle(ctx, &query.ListPets{
		Offset: offset,
		Limit:  limit,
	})
}
