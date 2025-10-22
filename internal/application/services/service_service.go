package services

import (
	"context"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
)

// ServiceService handles service operations
type ServiceService struct {
	createServiceHandler      *command.CreateServiceWithUoWHandler
	updateServiceHandler      *command.UpdateServiceWithUoWHandler
	deleteServiceHandler      *command.DeleteServiceWithUoWHandler
	getServiceHandler         *query.GetServiceHandler
	listVendorServicesHandler *query.ListVendorServicesHandler
	listServicesHandler       *query.ListServicesHandler
}

// NewServiceService creates a new service service
func NewServiceService(
	createServiceHandler *command.CreateServiceWithUoWHandler,
	updateServiceHandler *command.UpdateServiceWithUoWHandler,
	deleteServiceHandler *command.DeleteServiceWithUoWHandler,
	getServiceHandler *query.GetServiceHandler,
	listVendorServicesHandler *query.ListVendorServicesHandler,
	listServicesHandler *query.ListServicesHandler,
) *ServiceService {
	return &ServiceService{
		createServiceHandler:      createServiceHandler,
		updateServiceHandler:      updateServiceHandler,
		deleteServiceHandler:      deleteServiceHandler,
		getServiceHandler:         getServiceHandler,
		listVendorServicesHandler: listVendorServicesHandler,
		listServicesHandler:       listServicesHandler,
	}
}

// CreateService creates a new service
func (s *ServiceService) CreateService(ctx context.Context, cmd *command.CreateService) error {
	return s.createServiceHandler.Handle(ctx, cmd)
}

// UpdateService updates an existing service
func (s *ServiceService) UpdateService(ctx context.Context, cmd *command.UpdateService) error {
	return s.updateServiceHandler.Handle(ctx, cmd)
}

// DeleteService deletes a service
func (s *ServiceService) DeleteService(ctx context.Context, cmd *command.DeleteService) error {
	return s.deleteServiceHandler.Handle(ctx, cmd)
}

// GetService retrieves a service by ID
func (s *ServiceService) GetService(ctx context.Context, serviceID string) (interface{}, error) {
	return s.getServiceHandler.Handle(ctx, serviceID)
}

// ListVendorServices retrieves all services for a vendor with pagination
func (s *ServiceService) ListVendorServices(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
	return s.listVendorServicesHandler.Handle(ctx, vendorID, offset, limit)
}

// ListServices retrieves all services with pagination
func (s *ServiceService) ListServices(ctx context.Context, offset, limit int) ([]interface{}, error) {
	return s.listServicesHandler.Handle(ctx, offset, limit)
}
