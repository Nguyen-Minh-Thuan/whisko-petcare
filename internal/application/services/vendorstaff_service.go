package services

import (
	"context"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
)

// VendorStaffService handles vendor staff operations
type VendorStaffService struct {
	createVendorStaffHandler       *command.CreateVendorStaffWithUoWHandler
	deleteVendorStaffHandler       *command.DeleteVendorStaffWithUoWHandler
	getVendorStaffHandler          *query.GetVendorStaffHandler
	listVendorStaffByVendorHandler *query.ListVendorStaffByVendorHandler
	listVendorStaffByUserHandler   *query.ListVendorStaffByUserHandler
	listVendorStaffsHandler        *query.ListVendorStaffsHandler
}

// NewVendorStaffService creates a new vendor staff service
func NewVendorStaffService(
	createVendorStaffHandler *command.CreateVendorStaffWithUoWHandler,
	deleteVendorStaffHandler *command.DeleteVendorStaffWithUoWHandler,
	getVendorStaffHandler *query.GetVendorStaffHandler,
	listVendorStaffByVendorHandler *query.ListVendorStaffByVendorHandler,
	listVendorStaffByUserHandler *query.ListVendorStaffByUserHandler,
	listVendorStaffsHandler *query.ListVendorStaffsHandler,
) *VendorStaffService {
	return &VendorStaffService{
		createVendorStaffHandler:       createVendorStaffHandler,
		deleteVendorStaffHandler:       deleteVendorStaffHandler,
		getVendorStaffHandler:          getVendorStaffHandler,
		listVendorStaffByVendorHandler: listVendorStaffByVendorHandler,
		listVendorStaffByUserHandler:   listVendorStaffByUserHandler,
		listVendorStaffsHandler:        listVendorStaffsHandler,
	}
}

// CreateVendorStaff creates a new vendor staff
func (s *VendorStaffService) CreateVendorStaff(ctx context.Context, cmd *command.CreateVendorStaff) error {
	return s.createVendorStaffHandler.Handle(ctx, cmd)
}

// DeleteVendorStaff deletes a vendor staff
func (s *VendorStaffService) DeleteVendorStaff(ctx context.Context, cmd *command.DeleteVendorStaff) error {
	return s.deleteVendorStaffHandler.Handle(ctx, cmd)
}

// GetVendorStaff retrieves a vendor staff by user ID and vendor ID
func (s *VendorStaffService) GetVendorStaff(ctx context.Context, userID, vendorID string) (interface{}, error) {
	return s.getVendorStaffHandler.Handle(ctx, userID, vendorID)
}

// ListVendorStaffByVendor retrieves all vendor staffs for a vendor with pagination
func (s *VendorStaffService) ListVendorStaffByVendor(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
	return s.listVendorStaffByVendorHandler.Handle(ctx, vendorID, offset, limit)
}

// ListVendorStaffByUser retrieves all vendor staffs for a user with pagination
func (s *VendorStaffService) ListVendorStaffByUser(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	return s.listVendorStaffByUserHandler.Handle(ctx, userID, offset, limit)
}

// ListVendorStaffs retrieves all vendor staffs with pagination
func (s *VendorStaffService) ListVendorStaffs(ctx context.Context, offset, limit int) ([]interface{}, error) {
	return s.listVendorStaffsHandler.Handle(ctx, offset, limit)
}
