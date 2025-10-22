package query

import (
	"context"
	"fmt"

	"whisko-petcare/pkg/errors"
)

// VendorProjection interface for vendor read model
type VendorProjection interface {
	GetByID(ctx context.Context, id string) (interface{}, error)
	ListAll(ctx context.Context, offset, limit int) ([]interface{}, error)
}

// GetVendorHandler handles get vendor by ID queries
type GetVendorHandler struct {
	projection VendorProjection
}

// NewGetVendorHandler creates a new get vendor handler
func NewGetVendorHandler(projection VendorProjection) *GetVendorHandler {
	return &GetVendorHandler{
		projection: projection,
	}
}

// Handle processes the get vendor query
func (h *GetVendorHandler) Handle(ctx context.Context, vendorID string) (interface{}, error) {
	if vendorID == "" {
		return nil, errors.NewValidationError("vendor_id is required")
	}

	vendor, err := h.projection.GetByID(ctx, vendorID)
	if err != nil {
		return nil, errors.NewNotFoundError("vendor")
	}

	return vendor, nil
}

// ListVendorsHandler handles list vendors queries
type ListVendorsHandler struct {
	projection VendorProjection
}

// NewListVendorsHandler creates a new list vendors handler
func NewListVendorsHandler(projection VendorProjection) *ListVendorsHandler {
	return &ListVendorsHandler{
		projection: projection,
	}
}

// Handle processes the list vendors query
func (h *ListVendorsHandler) Handle(ctx context.Context, offset, limit int) ([]interface{}, error) {
	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	vendors, err := h.projection.ListAll(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list vendors: %v", err))
	}

	return vendors, nil
}
