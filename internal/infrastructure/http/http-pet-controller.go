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
	var cmd command.CreatePet
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	if err := c.petService.CreatePet(r.Context(), cmd); err != nil {
		if strings.Contains(err.Error(), "validation") {
			response.SendBadRequest(w, r, err.Error())
			return
		}
		response.SendInternalError(w, r, "Failed to create pet")
		return
	}

	response.SendCreated(w, r, map[string]string{"message": "Pet created successfully"})
}

// GetPet handles GET /pets/{id}
func (c *HTTPPetController) GetPet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	petID := strings.Split(path, "/")[0]
	
	if petID == "" {
		response.SendBadRequest(w, r, "Pet ID is required")
		return
	}

	pet, err := c.petService.GetPet(r.Context(), petID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Pet not found")
			return
		}
		response.SendInternalError(w, r, "Failed to get pet")
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
		response.SendInternalError(w, r, "Failed to list pets")
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
		response.SendBadRequest(w, r, "User ID is required")
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
		response.SendInternalError(w, r, "Failed to list user pets")
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
		response.SendBadRequest(w, r, "Pet ID is required")
		return
	}

	var cmd command.UpdatePet
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	cmd.PetID = petID

	if err := c.petService.UpdatePet(r.Context(), cmd); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Pet not found")
			return
		}
		if strings.Contains(err.Error(), "validation") {
			response.SendBadRequest(w, r, err.Error())
			return
		}
		response.SendInternalError(w, r, "Failed to update pet")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Pet updated successfully"})
}

// DeletePet handles DELETE /pets/{id}
func (c *HTTPPetController) DeletePet(w http.ResponseWriter, r *http.Request) {
	// Extract pet ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/pets/")
	petID := strings.Split(path, "/")[0]
	
	if petID == "" {
		response.SendBadRequest(w, r, "Pet ID is required")
		return
	}

	cmd := command.DeletePet{PetID: petID}

	if err := c.petService.DeletePet(r.Context(), cmd); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Pet not found")
			return
		}
		response.SendInternalError(w, r, "Failed to delete pet")
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Pet deleted successfully"})
}
