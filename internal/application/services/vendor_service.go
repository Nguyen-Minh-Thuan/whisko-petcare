package services

import (
	"context"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
)

// VendorService handles vendor operations
type VendorService struct {
	createVendorHandler       *command.CreateVendorWithUoWHandler
	updateVendorHandler       *command.UpdateVendorWithUoWHandler
	deleteVendorHandler       *command.DeleteVendorWithUoWHandler
	updateVendorImageHandler  *command.UpdateVendorImageWithUoWHandler
	getVendorHandler          *query.GetVendorHandler
	listVendorsHandler        *query.ListVendorsHandler
}

// NewVendorService creates a new vendor service
func NewVendorService(
	createVendorHandler *command.CreateVendorWithUoWHandler,
	updateVendorHandler *command.UpdateVendorWithUoWHandler,
	deleteVendorHandler *command.DeleteVendorWithUoWHandler,
	updateVendorImageHandler *command.UpdateVendorImageWithUoWHandler,
	getVendorHandler *query.GetVendorHandler,
	listVendorsHandler *query.ListVendorsHandler,
) *VendorService {
	return &VendorService{
		createVendorHandler:      createVendorHandler,
		updateVendorHandler:      updateVendorHandler,
		deleteVendorHandler:      deleteVendorHandler,
		updateVendorImageHandler: updateVendorImageHandler,
		getVendorHandler:         getVendorHandler,
		listVendorsHandler:       listVendorsHandler,
	}
}

// CreateVendor creates a new vendor
func (s *VendorService) CreateVendor(ctx context.Context, cmd *command.CreateVendor) error {
	return s.createVendorHandler.Handle(ctx, cmd)
}

// UpdateVendor updates an existing vendor
func (s *VendorService) UpdateVendor(ctx context.Context, cmd *command.UpdateVendor) error {
	return s.updateVendorHandler.Handle(ctx, cmd)
}

// DeleteVendor deletes a vendor
func (s *VendorService) DeleteVendor(ctx context.Context, cmd *command.DeleteVendor) error {
	return s.deleteVendorHandler.Handle(ctx, cmd)
}

// GetVendor retrieves a vendor by ID
func (s *VendorService) GetVendor(ctx context.Context, vendorID string) (interface{}, error) {
	return s.getVendorHandler.Handle(ctx, vendorID)
}

// ListVendors retrieves all vendors with pagination
func (s *VendorService) ListVendors(ctx context.Context, offset, limit int) ([]interface{}, error) {
	return s.listVendorsHandler.Handle(ctx, offset, limit)
}

// UpdateVendorImage updates a vendor's image URL
func (s *VendorService) UpdateVendorImage(ctx context.Context, cmd command.UpdateVendorImage) error {
	return s.updateVendorImageHandler.Handle(ctx, &cmd)
}
