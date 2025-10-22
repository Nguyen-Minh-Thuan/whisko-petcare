package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

// VendorController handles HTTP requests for vendor operations
type VendorController struct {
	service *services.VendorService
}

// NewVendorController creates a new vendor controller
func NewVendorController(service *services.VendorService) *VendorController {
	return &VendorController{
		service: service,
	}
}

// generateVendorID generates a unique vendor ID
func generateVendorID() string {
	return fmt.Sprintf("vendor_%d", time.Now().UnixNano())
}

// CreateVendor handles POST /vendors
func (c *VendorController) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email,omitempty"`
		Phone   string `json:"phone,omitempty"`
		Address string `json:"address,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CreateVendor{
		VendorID: generateVendorID(),
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := c.service.CreateVendor(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"id":      cmd.VendorID,
		"message": "Vendor created successfully",
	}
	response.SendCreated(w, r, responseData)
}

// GetVendor handles GET /vendors/{id}
func (c *VendorController) GetVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	vendor, err := c.service.GetVendor(r.Context(), vendorID)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, vendor)
}

// ListVendors handles GET /vendors
func (c *VendorController) ListVendors(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendors, err := c.service.ListVendors(r.Context(), offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, vendors)
}

// UpdateVendor handles PUT /vendors/{id}
func (c *VendorController) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	var req struct {
		Name    string `json:"name,omitempty"`
		Email   string `json:"email,omitempty"`
		Phone   string `json:"phone,omitempty"`
		Address string `json:"address,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.UpdateVendor{
		VendorID: vendorID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Address:  req.Address,
	}

	if err := c.service.UpdateVendor(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Vendor updated successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// DeleteVendor handles DELETE /vendors/{id}
func (c *VendorController) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	cmd := &command.DeleteVendor{
		VendorID: vendorID,
	}

	if err := c.service.DeleteVendor(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Vendor deleted successfully",
	}
	response.SendSuccess(w, r, responseData)
}
