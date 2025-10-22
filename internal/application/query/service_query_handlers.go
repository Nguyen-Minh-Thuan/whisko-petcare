package query

import (
	"context"
	"fmt"

	"whisko-petcare/pkg/errors"
)

// ServiceProjection interface for service read model
type ServiceProjection interface {
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error)
	ListAll(ctx context.Context, offset, limit int) ([]interface{}, error)
}

// GetServiceHandler handles get service by ID queries
type GetServiceHandler struct {
	projection ServiceProjection
}

// NewGetServiceHandler creates a new get service handler
func NewGetServiceHandler(projection ServiceProjection) *GetServiceHandler {
	return &GetServiceHandler{
		projection: projection,
	}
}

// Handle processes the get service query
func (h *GetServiceHandler) Handle(ctx context.Context, serviceID string) (interface{}, error) {
	if serviceID == "" {
		return nil, errors.NewValidationError("service_id is required")
	}

	service, err := h.projection.GetByID(ctx, serviceID)
	if err != nil {
		return nil, errors.NewNotFoundError("service")
	}

	return service, nil
}

// ListVendorServicesHandler handles list vendor services queries
type ListVendorServicesHandler struct {
	projection ServiceProjection
}

// NewListVendorServicesHandler creates a new list vendor services handler
func NewListVendorServicesHandler(projection ServiceProjection) *ListVendorServicesHandler {
	return &ListVendorServicesHandler{
		projection: projection,
	}
}

// Handle processes the list vendor services query
func (h *ListVendorServicesHandler) Handle(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
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

	services, err := h.projection.GetByVendorID(ctx, vendorID, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list vendor services: %v", err))
	}

	return services, nil
}

// ListServicesHandler handles list all services queries
type ListServicesHandler struct {
	projection ServiceProjection
}

// NewListServicesHandler creates a new list services handler
func NewListServicesHandler(projection ServiceProjection) *ListServicesHandler {
	return &ListServicesHandler{
		projection: projection,
	}
}

// Handle processes the list services query
func (h *ListServicesHandler) Handle(ctx context.Context, offset, limit int) ([]interface{}, error) {
	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	services, err := h.projection.ListAll(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list services: %v", err))
	}

	return services, nil
}
