package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
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

// CreateVendor handles POST /vendors
func (c *VendorController) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var cmd command.CreateVendor
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	if err := c.service.CreateVendor(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, "Failed to create vendor")
		return
	}

	response.SendCreated(w, r, map[string]string{"message": "Vendor created successfully"})
}

// GetVendor handles GET /vendors/{id}
func (c *VendorController) GetVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	vendor, err := c.service.GetVendor(r.Context(), vendorID)
	if err != nil {
		response.SendInternalError(w, r, "Failed to get vendor")
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
		response.SendInternalError(w, r, "Failed to list vendors")
		return
	}

	response.SendSuccess(w, r, vendors)
}

// UpdateVendor handles PUT /vendors/{id}
func (c *VendorController) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	var cmd command.UpdateVendor
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Set vendor ID from path
	cmd.VendorID = vendorID

	if err := c.service.UpdateVendor(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, "Failed to update vendor")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Vendor updated successfully"})
}

// DeleteVendor handles DELETE /vendors/{id}
func (c *VendorController) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("id")
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	cmd := &command.DeleteVendor{
		VendorID: vendorID,
	}

	if err := c.service.DeleteVendor(r.Context(), cmd); err != nil {
		response.SendInternalError(w, r, "Failed to delete vendor")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Vendor deleted successfully"})
}
