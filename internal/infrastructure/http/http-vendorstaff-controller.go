package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
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
	var req struct {
		UserID   string `json:"user_id"`
		VendorID string `json:"vendor_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CreateVendorStaff{
		UserID:   req.UserID,
		VendorID: req.VendorID,
	}

	if err := c.service.CreateVendorStaff(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"user_id":   cmd.UserID,
		"vendor_id": cmd.VendorID,
		"message":   "Vendor staff created successfully",
	}
	response.SendCreated(w, r, responseData)
}

// GetVendorStaff handles GET /vendor-staffs/{userID}/{vendorID}
func (c *VendorStaffController) GetVendorStaff(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and vendor ID from path
	userID := r.PathValue("userID")
	vendorID := r.PathValue("vendorID")
	
	if userID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("User ID is required"))
		return
	}
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	vendorStaff, err := c.service.GetVendorStaff(r.Context(), userID, vendorID)
	if err != nil {
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, vendorStaffs)
}

// ListVendorStaffByVendor handles GET /vendors/{vendorID}/staff
func (c *VendorStaffController) ListVendorStaffByVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := r.PathValue("vendorID")
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendorStaffs, err := c.service.ListVendorStaffByVendor(r.Context(), vendorID, offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, vendorStaffs)
}

// ListVendorStaffByUser handles GET /users/{userID}/vendor-staffs
func (c *VendorStaffController) ListVendorStaffByUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.PathValue("userID")
	if userID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("User ID is required"))
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	vendorStaffs, err := c.service.ListVendorStaffByUser(r.Context(), userID, offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, errors.NewValidationError("User ID is required"))
		return
	}
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}

	cmd := &command.DeleteVendorStaff{
		UserID:   userID,
		VendorID: vendorID,
	}

	if err := c.service.DeleteVendorStaff(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Vendor staff deleted successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// GetMyVendorProfile handles GET /vendor-staff/profile
// Returns the vendor staff profile with associated vendor information for the authenticated user
func (c *VendorStaffController) GetMyVendorProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		middleware.HandleError(w, r, errors.NewUnauthorizedError("Authentication required"))
		return
	}

	profile, err := c.service.GetVendorStaffProfile(r.Context(), userID)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, profile)
}
