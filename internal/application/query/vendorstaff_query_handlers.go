package query

import (
	"context"
	"fmt"

	"whisko-petcare/pkg/errors"
)

// VendorStaffProjection interface for vendor staff read model
type VendorStaffProjection interface {
	GetByID(ctx context.Context, userID, vendorID string) (interface{}, error)
	GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]interface{}, error)
	ListAll(ctx context.Context, offset, limit int) ([]interface{}, error)
}

// GetVendorStaffHandler handles get vendor staff by ID queries
type GetVendorStaffHandler struct {
	projection VendorStaffProjection
}

// NewGetVendorStaffHandler creates a new get vendor staff handler
func NewGetVendorStaffHandler(projection VendorStaffProjection) *GetVendorStaffHandler {
	return &GetVendorStaffHandler{
		projection: projection,
	}
}

// Handle processes the get vendor staff query
func (h *GetVendorStaffHandler) Handle(ctx context.Context, userID, vendorID string) (interface{}, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}
	if vendorID == "" {
		return nil, errors.NewValidationError("vendor_id is required")
	}

	vendorStaff, err := h.projection.GetByID(ctx, userID, vendorID)
	if err != nil {
		return nil, errors.NewNotFoundError("vendor staff")
	}

	return vendorStaff, nil
}

// ListVendorStaffByVendorHandler handles list vendor staff by vendor ID queries
type ListVendorStaffByVendorHandler struct {
	projection VendorStaffProjection
}

// NewListVendorStaffByVendorHandler creates a new list vendor staff by vendor handler
func NewListVendorStaffByVendorHandler(projection VendorStaffProjection) *ListVendorStaffByVendorHandler {
	return &ListVendorStaffByVendorHandler{
		projection: projection,
	}
}

// Handle processes the list vendor staff by vendor query
func (h *ListVendorStaffByVendorHandler) Handle(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
	if vendorID == "" {
		return nil, errors.NewValidationError("vendor_id is required")
	}

	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	vendorStaffs, err := h.projection.GetByVendorID(ctx, vendorID, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list vendor staff: %v", err))
	}

	return vendorStaffs, nil
}

// ListVendorStaffByUserHandler handles list vendor staff by user ID queries
type ListVendorStaffByUserHandler struct {
	projection VendorStaffProjection
}

// NewListVendorStaffByUserHandler creates a new list vendor staff by user handler
func NewListVendorStaffByUserHandler(projection VendorStaffProjection) *ListVendorStaffByUserHandler {
	return &ListVendorStaffByUserHandler{
		projection: projection,
	}
}

// Handle processes the list vendor staff by user query
func (h *ListVendorStaffByUserHandler) Handle(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}

	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	vendorStaffs, err := h.projection.GetByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list vendor staff: %v", err))
	}

	return vendorStaffs, nil
}

// ListVendorStaffsHandler handles list all vendor staffs queries
type ListVendorStaffsHandler struct {
	projection VendorStaffProjection
}

// NewListVendorStaffsHandler creates a new list vendor staffs handler
func NewListVendorStaffsHandler(projection VendorStaffProjection) *ListVendorStaffsHandler {
	return &ListVendorStaffsHandler{
		projection: projection,
	}
}

// Handle processes the list vendor staffs query
func (h *ListVendorStaffsHandler) Handle(ctx context.Context, offset, limit int) ([]interface{}, error) {
	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	vendorStaffs, err := h.projection.ListAll(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list vendor staffs: %v", err))
	}

	return vendorStaffs, nil
}
