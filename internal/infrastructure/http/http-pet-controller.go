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

// HTTPPetController handles HTTP requests for pet operations
type HTTPPetController struct {
	petService *services.PetService
	cloudinary *cloudinary.Service
}

// NewHTTPPetController creates a new HTTP pet controller
func NewHTTPPetController(petService *services.PetService, cloudinary *cloudinary.Service) *HTTPPetController {
	return &HTTPPetController{
		petService: petService,
		cloudinary: cloudinary,
	}
}

// CreatePet handles POST /pets - supports both JSON and multipart/form-data with image
func (c *HTTPPetController) CreatePet(w http.ResponseWriter, r *http.Request) {
	var userID, name, species, breed, imageUrl string
	var age int
	var weight float64
	petID := fmt.Sprintf("pet_%d", time.Now().UnixNano())

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get form fields
		userID = r.FormValue("user_id")
		name = r.FormValue("name")
		species = r.FormValue("species")
		breed = r.FormValue("breed")
		
		if ageStr := r.FormValue("age"); ageStr != "" {
			age, _ = strconv.Atoi(ageStr)
		}
		if weightStr := r.FormValue("weight"); weightStr != "" {
			weight, _ = strconv.ParseFloat(weightStr, 64)
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
			
			uploadRes, err := c.cloudinary.UploadPetImage(r.Context(), file, fileHeader.Filename, petID)
			if err != nil {
				middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
				return
			}
			imageUrl = uploadRes.SecureURL
		}
	} else {
		// JSON body
		var req struct {
			UserID   string  `json:"user_id"`
			Name     string  `json:"name"`
			Species  string  `json:"species"`
			Breed    string  `json:"breed,omitempty"`
			Age      int     `json:"age,omitempty"`
			Weight   float64 `json:"weight,omitempty"`
			ImageUrl string  `json:"image_url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		userID = req.UserID
		name = req.Name
		species = req.Species
		breed = req.Breed
		age = req.Age
		weight = req.Weight
		imageUrl = req.ImageUrl
	}

	cmd := command.CreatePet{
		UserID:   userID,
		Name:     name,
		Species:  species,
		Breed:    breed,
		Age:      age,
		Weight:   weight,
		ImageUrl: imageUrl,
	}

	if err := c.petService.CreatePet(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Pet created successfully",
		"pet_id":  petID,
	}
	if imageUrl != "" {
		responseData["image_url"] = imageUrl
	}
	response.SendCreated(w, r, responseData)
}

// GetPet handles GET /pets/{id}
func (c *HTTPPetController) GetPet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	petID := strings.Split(path, "/")[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	pet, err := c.petService.GetPet(r.Context(), petID)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, pet)
}

// ListPets handles GET /pets
func (c *HTTPPetController) ListPets(w http.ResponseWriter, r *http.Request) {
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
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	pets, err := c.petService.ListPets(r.Context(), offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"pets":   pets,
		"offset": offset,
		"limit":  limit,
		"count":  len(pets),
	}

	response.SendSuccess(w, r, responseData)
}

// ListUserPets handles GET /users/{userID}/pets
func (c *HTTPPetController) ListUserPets(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("User ID is required"))
		return
	}
	userID := parts[0]

	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 10

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	pets, err := c.petService.ListUserPets(r.Context(), userID, offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"pets":   pets,
		"offset": offset,
		"limit":  limit,
		"count":  len(pets),
	}

	response.SendSuccess(w, r, responseData)
}

// UpdatePet handles PUT /pets/{id}
func (c *HTTPPetController) UpdatePet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	petID := strings.Split(path, "/")[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	var req struct {
		Name    string  `json:"name,omitempty"`
		Species string  `json:"species,omitempty"`
		Breed   string  `json:"breed,omitempty"`
		Age     int     `json:"age,omitempty"`
		Weight  float64 `json:"weight,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.UpdatePet{
		PetID:   petID,
		Name:    req.Name,
		Species: req.Species,
		Breed:   req.Breed,
		Age:     req.Age,
		Weight:  req.Weight,
	}

	if err := c.petService.UpdatePet(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Pet updated successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// DeletePet handles DELETE /pets/{id}
func (c *HTTPPetController) DeletePet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	petID := strings.Split(path, "/")[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	cmd := command.DeletePet{PetID: petID}

	if err := c.petService.DeletePet(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Pet deleted successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// AddPetVaccination handles POST /pets/{id}/vaccinations
func (c *HTTPPetController) AddPetVaccination(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	parts := strings.Split(path, "/")
	petID := parts[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	var req struct {
		VaccineName  string `json:"vaccine_name"`
		Date         string `json:"date"`
		NextDueDate  string `json:"next_due_date,omitempty"`
		Veterinarian string `json:"veterinarian,omitempty"`
		Notes        string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	// Parse dates
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid date format, use RFC3339"))
		return
	}

	var nextDueDate time.Time
	if req.NextDueDate != "" {
		nextDueDate, err = time.Parse(time.RFC3339, req.NextDueDate)
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid next_due_date format, use RFC3339"))
			return
		}
	}

	cmd := command.AddPetVaccination{
		PetID:        petID,
		VaccineName:  req.VaccineName,
		Date:         date,
		NextDueDate:  nextDueDate,
		Veterinarian: req.Veterinarian,
		Notes:        req.Notes,
	}

	if err := c.petService.AddPetVaccination(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Vaccination record added successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// AddPetMedicalRecord handles POST /pets/{id}/medical-records
func (c *HTTPPetController) AddPetMedicalRecord(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	parts := strings.Split(path, "/")
	petID := parts[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	var req struct {
		Date         string `json:"date"`
		Description  string `json:"description"`
		Treatment    string `json:"treatment,omitempty"`
		Veterinarian string `json:"veterinarian,omitempty"`
		Diagnosis    string `json:"diagnosis,omitempty"`
		Notes        string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid date format, use RFC3339"))
		return
	}

	cmd := command.AddPetMedicalRecord{
		PetID:        petID,
		Date:         date,
		Description:  req.Description,
		Treatment:    req.Treatment,
		Veterinarian: req.Veterinarian,
		Diagnosis:    req.Diagnosis,
		Notes:        req.Notes,
	}

	if err := c.petService.AddPetMedicalRecord(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Medical record added successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// AddPetAllergy handles POST /pets/{id}/allergies
func (c *HTTPPetController) AddPetAllergy(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	parts := strings.Split(path, "/")
	petID := parts[0]
	
	if petID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}

	var req struct {
		Allergen      string `json:"allergen"`
		Severity      string `json:"severity"`
		Symptoms      string `json:"symptoms,omitempty"`
		DiagnosedDate string `json:"diagnosed_date,omitempty"`
		Notes         string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	// Parse diagnosed date if provided
	var diagnosedDate time.Time
	if req.DiagnosedDate != "" {
		var err error
		diagnosedDate, err = time.Parse(time.RFC3339, req.DiagnosedDate)
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid diagnosed_date format, use RFC3339"))
			return
		}
	}

	cmd := command.AddPetAllergy{
		PetID:         petID,
		Allergen:      req.Allergen,
		Severity:      req.Severity,
		Symptoms:      req.Symptoms,
		DiagnosedDate: diagnosedDate,
		Notes:         req.Notes,
	}

	if err := c.petService.AddPetAllergy(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Allergy added successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// RemovePetAllergy handles DELETE /pets/{id}/allergies/{allergy_id}
func (c *HTTPPetController) RemovePetAllergy(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID and allergy ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 3 {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID and Allergy ID are required"))
		return
	}
	
	petID := parts[0]
	allergyID := parts[2]
	
	if petID == "" || allergyID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID and Allergy ID are required"))
		return
	}

	cmd := command.RemovePetAllergy{
		PetID:     petID,
		AllergyID: allergyID,
	}

	if err := c.petService.RemovePetAllergy(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Allergy removed successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// UpdatePetImage handles PUT /pets/{id}/image - supports multipart/form-data with image file
func (c *HTTPPetController) UpdatePetImage(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Pet ID is required"))
		return
	}
	
	petID := parts[0]
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
		
		uploadRes, err := c.cloudinary.UploadPetImage(r.Context(), file, fileHeader.Filename, petID)
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

	cmd := command.UpdatePetImage{
		PetID:    petID,
		ImageUrl: imageUrl,
	}

	if err := c.petService.UpdatePetImage(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"message":   "Pet image updated successfully",
		"image_url": imageUrl,
	})
}
