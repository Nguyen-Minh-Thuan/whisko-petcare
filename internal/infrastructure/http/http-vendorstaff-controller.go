package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/response"
)

// VendorStaffController handles HTTP requests for vendor staff operations
type VendorStaffController struct {
	service *services.VendorStaffService
}

// NewVendorStaffController creates a new vendor staff controller
func NewVendorStaffController(service *services.VendorStaffService) *VendorStaffController {
	return &VendorStaffController{
		service: service,
	}
}

// CreateVendorStaff handles POST /vendor-staffs
func (c *VendorStaffController) CreateVendorStaff(w http.ResponseWriter, r *http.Request) {
	var cmd command.CreateVendorStaff
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	if err := c.service.CreateVendorStaff(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, "Failed to create vendor staff")
		return
	}

	response.SendCreated(w, r, map[string]string{"message": "Vendor staff created successfully"})
}

// GetVendorStaff handles GET /vendor-staffs/{userID}/{vendorID}
func (c *VendorStaffController) GetVendorStaff(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and vendor ID from path
	userID := r.PathValue("userID")
	vendorID := r.PathValue("vendorID")
	
	if userID == "" {
		response.SendBadRequest(w, r, "User ID is required")
		return
	}
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	vendorStaff, err := c.service.GetVendorStaff(r.Context(), userID, vendorID)
	if err != nil {
		response.SendInternalError(w, r, "Failed to get vendor staff")
		return
	}

	response.SendSuccess(w, r, vendorStaff)
}

// ListVendorStaffs handles GET /vendor-staffs
func (c *VendorStaffController) ListVendorStaffs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendorStaffs, err := c.service.ListVendorStaffs(r.Context(), offset, limit)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list vendor staffs")
		return
	}

	response.SendSuccess(w, r, vendorStaffs)
}

// ListVendorStaffByVendor handles GET /vendors/{vendorID}/staff
func (c *VendorStaffController) ListVendorStaffByVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("vendorID")
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendorStaffs, err := c.service.ListVendorStaffByVendor(r.Context(), vendorID, offset, limit)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list vendor staffs by vendor")
		return
	}

	response.SendSuccess(w, r, vendorStaffs)
}

// ListVendorStaffByUser handles GET /users/{userID}/vendor-staffs
func (c *VendorStaffController) ListVendorStaffByUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.PathValue("userID")
	if userID == "" {
		response.SendBadRequest(w, r, "User ID is required")
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendorStaffs, err := c.service.ListVendorStaffByUser(r.Context(), userID, offset, limit)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list vendor staffs by user")
		return
	}

	response.SendSuccess(w, r, vendorStaffs)
}

// DeleteVendorStaff handles DELETE /vendor-staffs/{userID}/{vendorID}
func (c *VendorStaffController) DeleteVendorStaff(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and vendor ID from path
	userID := r.PathValue("userID")
	vendorID := r.PathValue("vendorID")
	
	if userID == "" {
		response.SendBadRequest(w, r, "User ID is required")
		return
	}
	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	cmd := &command.DeleteVendorStaff{
		UserID:   userID,
		VendorID: vendorID,
	}

	if err := c.service.DeleteVendorStaff(r.Context(), cmd); err != nil {
		response.SendInternalError(w, r, "Failed to delete vendor staff")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Vendor staff deleted successfully"})
}
