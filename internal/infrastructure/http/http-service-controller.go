package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/response"
)

// HTTPServiceController handles HTTP requests for service operations
type HTTPServiceController struct {
	service *services.ServiceService
}

// NewHTTPServiceController creates a new HTTP service controller
func NewHTTPServiceController(service *services.ServiceService) *HTTPServiceController {
	return &HTTPServiceController{
		service: service,
	}
}

// CreateService handles POST /services
func (c *HTTPServiceController) CreateService(w http.ResponseWriter, r *http.Request) {
	var cmd command.CreateService
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	if err := c.service.CreateService(r.Context(), &cmd); err != nil {
		if strings.Contains(err.Error(), "validation") {
			response.SendBadRequest(w, r, err.Error())
			return
		}
		response.SendInternalError(w, r, "Failed to create service")
		return
	}

	response.SendCreated(w, r, map[string]string{"message": "Service created successfully"})
}

// GetService handles GET /services/{id}
func (c *HTTPServiceController) GetService(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	serviceID := strings.Split(path, "/")[0]
	
	if serviceID == "" {
		response.SendBadRequest(w, r, "Service ID is required")
		return
	}

	service, err := c.service.GetService(r.Context(), serviceID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Service not found")
			return
		}
		response.SendInternalError(w, r, "Failed to get service")
		return
	}

	response.SendSuccess(w, r, service)
}

// ListServices handles GET /services
func (c *HTTPServiceController) ListServices(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 10 // default limit

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	services, err := c.service.ListServices(r.Context(), offset, limit)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list services")
		return
	}

	response.SendSuccess(w, r, services)
}

// ListVendorServices handles GET /vendors/{id}/services
func (c *HTTPServiceController) ListVendorServices(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	vendorID := strings.Split(path, "/")[0]
	
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 10 // default limit

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	services, err := c.service.ListVendorServices(r.Context(), vendorID, offset, limit)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list vendor services")
		return
	}

	response.SendSuccess(w, r, services)
}

// UpdateService handles PUT /services/{id}
func (c *HTTPServiceController) UpdateService(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	serviceID := strings.Split(path, "/")[0]
	
	if serviceID == "" {
		response.SendBadRequest(w, r, "Service ID is required")
		return
	}

	var cmd command.UpdateService
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}
	cmd.ServiceID = serviceID

	if err := c.service.UpdateService(r.Context(), &cmd); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Service not found")
			return
		}
		if strings.Contains(err.Error(), "validation") {
			response.SendBadRequest(w, r, err.Error())
			return
		}
		response.SendInternalError(w, r, "Failed to update service")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Service updated successfully"})
}

// DeleteService handles DELETE /services/{id}
func (c *HTTPServiceController) DeleteService(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	serviceID := strings.Split(path, "/")[0]
	
	if serviceID == "" {
		response.SendBadRequest(w, r, "Service ID is required")
		return
	}

	cmd := command.DeleteService{ServiceID: serviceID}
	if err := c.service.DeleteService(r.Context(), &cmd); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Service not found")
			return
		}
		response.SendInternalError(w, r, "Failed to delete service")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Service deleted successfully"})
}
