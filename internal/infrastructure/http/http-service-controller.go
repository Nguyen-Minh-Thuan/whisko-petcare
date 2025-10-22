package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
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
	var req struct {
		VendorID    string `json:"vendor_id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Price       int    `json:"price"`
		Duration    int    `json:"duration_minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CreateService{
		VendorID:    req.VendorID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
	}

	if err := c.service.CreateService(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Service created successfully",
	}
	response.SendCreated(w, r, responseData)
}

// GetService handles GET /services/{id}
func (c *HTTPServiceController) GetService(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	serviceID := strings.Split(path, "/")[0]
	
	if serviceID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Service ID is required"))
		return
	}

	service, err := c.service.GetService(r.Context(), serviceID)
	if err != nil {
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
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
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, errors.NewValidationError("Service ID is required"))
		return
	}

	var req struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Price       int    `json:"price,omitempty"`
		Duration    int    `json:"duration_minutes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.UpdateService{
		ServiceID:   serviceID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
	}

	if err := c.service.UpdateService(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Service updated successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// DeleteService handles DELETE /services/{id}
func (c *HTTPServiceController) DeleteService(w http.ResponseWriter, r *http.Request) {
	// Extract service ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/services/")
	serviceID := strings.Split(path, "/")[0]
	
	if serviceID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Service ID is required"))
		return
	}

	cmd := command.DeleteService{ServiceID: serviceID}
	if err := c.service.DeleteService(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Service deleted successfully",
	}
	response.SendSuccess(w, r, responseData)
}
