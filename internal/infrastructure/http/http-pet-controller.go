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

// HTTPPetController handles HTTP requests for pet operations
type HTTPPetController struct {
	petService *services.PetService
}

// NewHTTPPetController creates a new HTTP pet controller
func NewHTTPPetController(petService *services.PetService) *HTTPPetController {
	return &HTTPPetController{
		petService: petService,
	}
}

// CreatePet handles POST /pets
func (c *HTTPPetController) CreatePet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string  `json:"user_id"`
		Name    string  `json:"name"`
		Species string  `json:"species"`
		Breed   string  `json:"breed,omitempty"`
		Age     int     `json:"age,omitempty"`
		Weight  float64 `json:"weight,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CreatePet{
		UserID:  req.UserID,
		Name:    req.Name,
		Species: req.Species,
		Breed:   req.Breed,
		Age:     req.Age,
		Weight:  req.Weight,
	}

	if err := c.petService.CreatePet(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Pet created successfully",
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
