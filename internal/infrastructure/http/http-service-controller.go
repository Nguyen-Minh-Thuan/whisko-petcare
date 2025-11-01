package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/internal/infrastructure/cloudinary"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

// HTTPServiceController handles HTTP requests for service operations
type HTTPServiceController struct {
	service    *services.ServiceService
	cloudinary *cloudinary.Service
}

// NewHTTPServiceController creates a new HTTP service controller
func NewHTTPServiceController(service *services.ServiceService, cloudinary *cloudinary.Service) *HTTPServiceController {
	return &HTTPServiceController{
		service:    service,
		cloudinary: cloudinary,
	}
}

// CreateService handles POST /services - supports both JSON and multipart/form-data with image
func (c *HTTPServiceController) CreateService(w http.ResponseWriter, r *http.Request) {
	var vendorID, name, description, imageUrl string
	var price, duration int
	var tags []string
	serviceID := fmt.Sprintf("service_%d", time.Now().UnixNano())

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get form fields
		vendorID = r.FormValue("vendor_id")
		name = r.FormValue("name")
		description = r.FormValue("description")
		
		if priceStr := r.FormValue("price"); priceStr != "" {
			price, _ = strconv.Atoi(priceStr)
		}
		if durationStr := r.FormValue("duration_minutes"); durationStr != "" {
			duration, _ = strconv.Atoi(durationStr)
		}
		
		// Parse tags if provided (comma-separated)
		if tagsStr := r.FormValue("tags"); tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		// Check if image file is provided
		file, fileHeader, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			
			// Upload to Cloudinary
			if c.cloudinary == nil {
				middleware.HandleError(w, r, errors.NewInternalError("Cloudinary not configured"))
				return
			}
			
			uploadRes, err := c.cloudinary.UploadServiceImage(r.Context(), file, fileHeader.Filename, serviceID)
			if err != nil {
				middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
				return
			}
			imageUrl = uploadRes.SecureURL
		}
	} else {
		// JSON body
		var req struct {
			VendorID    string   `json:"vendor_id"`
			Name        string   `json:"name"`
			Description string   `json:"description,omitempty"`
			Price       int      `json:"price"`
			Duration    int      `json:"duration_minutes"`
			Tags        []string `json:"tags,omitempty"`
			ImageUrl    string   `json:"image_url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		vendorID = req.VendorID
		name = req.Name
		description = req.Description
		price = req.Price
		duration = req.Duration
		tags = req.Tags
		imageUrl = req.ImageUrl
	}

	cmd := command.CreateService{
		VendorID:    vendorID,
		Name:        name,
		Description: description,
		Price:       price,
		Duration:    duration,
		Tags:        tags,
		ImageUrl:    imageUrl,
	}

	if err := c.service.CreateService(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message":    "Service created successfully",
		"service_id": serviceID,
	}
	if imageUrl != "" {
		responseData["image_url"] = imageUrl
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

// ListVendorServices handles GET /vendors/{id}/services and GET /api/services/vendor/{id}
func (c *HTTPServiceController) ListVendorServices(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from URL path - supports both patterns
	var vendorID string
	
	if strings.HasPrefix(r.URL.Path, "/api/services/vendor/") {
		// Pattern: /api/services/vendor/{vendorID}
		path := strings.TrimPrefix(r.URL.Path, "/api/services/vendor/")
		vendorID = strings.Split(path, "/")[0]
	} else {
		// Pattern: /vendors/{vendorID}/services
		path := strings.TrimPrefix(r.URL.Path, "/vendors/")
		vendorID = strings.Split(path, "/")[0]
	}
	
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
		Name        string   `json:"name,omitempty"`
		Description string   `json:"description,omitempty"`
		Price       int      `json:"price,omitempty"`
		Duration    int      `json:"duration_minutes,omitempty"`
		Tags        []string `json:"tags,omitempty"`
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
		Tags:        req.Tags,
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
