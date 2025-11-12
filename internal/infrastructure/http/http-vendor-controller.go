package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/internal/infrastructure/cloudinary"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"

	"github.com/google/uuid"
)

// VendorController handles HTTP requests for vendor operations
type VendorController struct {
	service    *services.VendorService
	cloudinary *cloudinary.Service
}

// NewVendorController creates a new vendor controller
func NewVendorController(service *services.VendorService, cloudinary *cloudinary.Service) *VendorController {
	return &VendorController{
		service:    service,
		cloudinary: cloudinary,
	}
}

// generateVendorID generates a unique vendor ID using UUID
func generateVendorID() string {
	return uuid.New().String()
}

// CreateVendor handles POST /vendors - supports both JSON and multipart/form-data with image
func (c *VendorController) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var name, email, phone, address, imageUrl string
	vendorID := generateVendorID()

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get form fields
		name = r.FormValue("name")
		email = r.FormValue("email")
		phone = r.FormValue("phone")
		address = r.FormValue("address")

		// Check if image file is provided
		file, fileHeader, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			
			// Upload to Cloudinary
			if c.cloudinary == nil {
				middleware.HandleError(w, r, errors.NewInternalError("Cloudinary not configured"))
				return
			}
			
			uploadRes, err := c.cloudinary.UploadVendorImage(r.Context(), file, fileHeader.Filename, vendorID)
			if err != nil {
				middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
				return
			}
			imageUrl = uploadRes.SecureURL
		}
	} else {
		// JSON body
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email,omitempty"`
			Phone    string `json:"phone,omitempty"`
			Address  string `json:"address,omitempty"`
			ImageUrl string `json:"image_url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		name = req.Name
		email = req.Email
		phone = req.Phone
		address = req.Address
		imageUrl = req.ImageUrl
	}

	cmd := command.CreateVendor{
		VendorID: vendorID,
		Name:     name,
		Email:    email,
		Phone:    phone,
		Address:  address,
		ImageUrl: imageUrl,
	}

	if err := c.service.CreateVendor(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"id":      cmd.VendorID,
		"message": "Vendor created successfully",
	}
	if imageUrl != "" {
		responseData["image_url"] = imageUrl
	}
	response.SendCreated(w, r, responseData)
}

// GetVendor handles GET /vendors/{id}
func (c *VendorController) GetVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path using manual parsing (not PathValue - doesn't work with prefix routes)
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	vendorID := ""
	if len(parts) > 0 && parts[0] != "" {
		vendorID = parts[0]
	}
	
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
	// Extract vendor ID from path using manual parsing (not PathValue - doesn't work with prefix routes)
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	vendorID := ""
	if len(parts) > 0 && parts[0] != "" {
		vendorID = parts[0]
	}
	
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
	// Extract vendor ID from path using manual parsing (not PathValue - doesn't work with prefix routes)
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	vendorID := ""
	if len(parts) > 0 && parts[0] != "" {
		vendorID = parts[0]
	}
	
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

// UpdateVendorImage handles PUT /vendors/{id}/image - supports multipart/form-data with image file
func (c *VendorController) UpdateVendorImage(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}
	
	vendorID := parts[0]
	var imageUrl string

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get image file
		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Image file is required"))
			return
		}
		defer file.Close()
		
		// Upload to Cloudinary
		if c.cloudinary == nil {
			middleware.HandleError(w, r, errors.NewInternalError("Cloudinary not configured"))
			return
		}
		
		uploadRes, err := c.cloudinary.UploadVendorImage(r.Context(), file, fileHeader.Filename, vendorID)
		if err != nil {
			middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
			return
		}
		imageUrl = uploadRes.SecureURL
	} else {
		// JSON body (backward compatibility)
		var req struct {
			ImageUrl string `json:"image_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		if req.ImageUrl == "" {
			middleware.HandleError(w, r, errors.NewValidationError("image_url is required"))
			return
		}
		imageUrl = req.ImageUrl
	}

	cmd := command.UpdateVendorImage{
		VendorID: vendorID,
		ImageUrl: imageUrl,
	}

	if err := c.service.UpdateVendorImage(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"message":   "Vendor image updated successfully",
		"image_url": imageUrl,
	})
}

// UpdateBankAccount handles PUT /vendors/{id}/bank-account - Update vendor's bank account for payouts
func (c *VendorController) UpdateBankAccount(w http.ResponseWriter, r *http.Request) {
	// Debug logging - request info
	fmt.Printf("\n🔍 === UpdateBankAccount DEBUG START ===\n")
	fmt.Printf("📍 Request URL: %s\n", r.URL.Path)
	fmt.Printf("📍 Request Method: %s\n", r.Method)
	fmt.Printf("📍 Content-Type: %s\n", r.Header.Get("Content-Type"))
	
	// Extract vendor ID from path using PathValue (Go 1.22+)
	vendorID := r.PathValue("id")
	
	fmt.Printf("🔑 Extracted vendorID (PathValue): '%s'\n", vendorID)
	fmt.Printf("🔑 vendorID length: %d\n", len(vendorID))
	fmt.Printf("🔑 vendorID is empty: %v\n", vendorID == "")
	
	// Fallback: manual path parsing for debugging
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	fmt.Printf("📂 Path after trim: '%s'\n", path)
	fmt.Printf("📂 Path parts: %v\n", parts)
	fmt.Printf("📂 Parts length: %d\n", len(parts))
	if len(parts) > 0 {
		fmt.Printf("📂 parts[0]: '%s'\n", parts[0])
	}
	
	// Try both methods
	if vendorID == "" && len(parts) > 0 && parts[0] != "" {
		vendorID = parts[0]
		fmt.Printf("⚠️  Using fallback vendor ID from path parsing: '%s'\n", vendorID)
	}
	
	if vendorID == "" {
		fmt.Printf("❌ VALIDATION FAILED: Vendor ID is empty after all extraction attempts\n")
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, errors.NewValidationError("Vendor ID is required"))
		return
	}
	
	fmt.Printf("✅ Vendor ID validated: '%s'\n", vendorID)
	
	// Parse request body
	var req struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		BankBranch    string `json:"bank_branch"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ JSON decode error: %v\n", err)
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}
	
	fmt.Printf("📄 Request body parsed:\n")
	fmt.Printf("   BankName: '%s'\n", req.BankName)
	fmt.Printf("   AccountNumber: '%s'\n", req.AccountNumber)
	fmt.Printf("   AccountName: '%s'\n", req.AccountName)
	fmt.Printf("   BankBranch: '%s'\n", req.BankBranch)
	
	// Validate required fields
	if req.BankName == "" {
		fmt.Printf("❌ VALIDATION FAILED: BankName is empty\n")
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, errors.NewValidationError("Bank name is required"))
		return
	}
	if req.AccountNumber == "" {
		fmt.Printf("❌ VALIDATION FAILED: AccountNumber is empty\n")
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, errors.NewValidationError("Account number is required"))
		return
	}
	if req.AccountName == "" {
		fmt.Printf("❌ VALIDATION FAILED: AccountName is empty\n")
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, errors.NewValidationError("Account name is required"))
		return
	}
	
	fmt.Printf("✅ All validations passed\n")
	
	// Create command
	cmd := command.UpdateVendorBankAccount{
		VendorID:      vendorID,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		BankBranch:    req.BankBranch,
	}
	
	fmt.Printf("📤 Calling service.UpdateVendorBankAccount...\n")
	
	// Call service
	if err := c.service.UpdateVendorBankAccount(r.Context(), cmd); err != nil {
		fmt.Printf("❌ Service error: %v\n", err)
		fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
		middleware.HandleError(w, r, err)
		return
	}
	
	fmt.Printf("✅ Service call successful\n")
	fmt.Printf("🔍 === UpdateBankAccount DEBUG END ===\n\n")
	
	// Return success response
	response.SendSuccess(w, r, map[string]interface{}{
		"message": "Bank account updated successfully",
		"vendor_id": vendorID,
		"bank_account": map[string]interface{}{
			"bank_name":      req.BankName,
			"account_number": req.AccountNumber,
			"account_name":   req.AccountName,
			"bank_branch":    req.BankBranch,
		},
	})
}
