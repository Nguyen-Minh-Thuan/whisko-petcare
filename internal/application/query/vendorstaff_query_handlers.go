package query

import (
	"context"
	"fmt"

	"whisko-petcare/internal/infrastructure/projection"
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

// GetVendorStaffProfileHandler handles getting vendor staff profile with vendor info
type GetVendorStaffProfileHandler struct {
	vendorStaffProjection VendorStaffProjection
	vendorProjection      VendorProjection
}

// NewGetVendorStaffProfileHandler creates a new get vendor staff profile handler
func NewGetVendorStaffProfileHandler(
	vendorStaffProjection VendorStaffProjection,
	vendorProjection VendorProjection,
) *GetVendorStaffProfileHandler {
	return &GetVendorStaffProfileHandler{
		vendorStaffProjection: vendorStaffProjection,
		vendorProjection:      vendorProjection,
	}
}

// Handle processes the get vendor staff profile query
func (h *GetVendorStaffProfileHandler) Handle(ctx context.Context, userID string) (interface{}, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}

	// Get all vendor staff entries for this user
	vendorStaffs, err := h.vendorStaffProjection.GetByUserID(ctx, userID, 0, 100)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to get vendor staffs: %v", err))
	}

	if len(vendorStaffs) == 0 {
		return nil, errors.NewNotFoundError("vendor staff profile")
	}

	// Prepare response with vendor staff info and associated vendor details
	var profiles []interface{}
	
	for _, staffInterface := range vendorStaffs {
		// Cast to VendorStaffReadModel (not map)
		staff, ok := staffInterface.(projection.VendorStaffReadModel)
		if !ok {
			continue
		}

		// Get vendor information
		vendorInterface, err := h.vendorProjection.GetByID(ctx, staff.VendorID)
		if err != nil {
			// If vendor not found, skip this entry but don't fail entirely
			continue
		}

		// Cast vendor to VendorReadModel (it's returned as a value, not pointer)
		vendor, ok := vendorInterface.(projection.VendorReadModel)
		if !ok {
			continue
		}

		// Combine staff and vendor information
		profile := map[string]interface{}{
			"staff_id":   staff.ID,
			"vendor_id":  staff.VendorID,
			"user_id":    staff.UserID,
			"is_active":  staff.IsActive,
			"created_at": staff.CreatedAt,
			"updated_at": staff.UpdatedAt,
			"vendor_name":    vendor.Name,
			"vendor_email":   vendor.Email,
			"vendor_phone":   vendor.Phone,
			"vendor_address": vendor.Address,
		}
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 {
		return nil, errors.NewNotFoundError("valid vendor staff profile")
	}

	// If there's only one profile, return it directly; otherwise return array
	if len(profiles) == 1 {
		return profiles[0], nil
	}

	return profiles, nil
}
