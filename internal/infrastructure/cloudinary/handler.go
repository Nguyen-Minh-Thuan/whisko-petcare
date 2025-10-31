package cloudinary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"whisko-petcare/pkg/response"
)

// Handler provides HTTP handlers for Cloudinary operations
type Handler struct {
	service *Service
}

// NewHandler creates a new Cloudinary HTTP handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// UploadImageRequest represents the upload image request
type UploadImageRequest struct {
	EntityType string `json:"entity_type"` // "pet", "vendor", "user", "general"
	EntityID   string `json:"entity_id"`   // ID of the entity (optional for general uploads)
	Folder     string `json:"folder"`      // Custom folder (optional)
}

// UploadImageResponse represents the upload image response
type UploadImageResponse struct {
	PublicID  string `json:"public_id"`
	SecureURL string `json:"secure_url"`
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
}

// HandleUploadImage handles image upload requests
func (h *Handler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.SendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.SendError(w, r, http.StatusBadRequest, "INVALID_FORM", "Failed to parse multipart form")
		return
	}

	// Get the file from the form
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		response.SendError(w, r, http.StatusBadRequest, "MISSING_FILE", "Image file is required")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		response.SendError(w, r, http.StatusBadRequest, "INVALID_FILE_TYPE", "File must be an image")
		return
	}

	// Parse request parameters
	entityType := r.FormValue("entity_type")
	entityID := r.FormValue("entity_id")
	folder := r.FormValue("folder")

	var result *UploadResult

	// Upload based on entity type
	ctx := r.Context()
	switch entityType {
	case "pet":
		if entityID == "" {
			response.SendError(w, r, http.StatusBadRequest, "MISSING_ENTITY_ID", "Entity ID is required for pet uploads")
			return
		}
		result, err = h.service.UploadPetImage(ctx, file, fileHeader.Filename, entityID)

	case "vendor":
		if entityID == "" {
			response.SendError(w, r, http.StatusBadRequest, "MISSING_ENTITY_ID", "Entity ID is required for vendor uploads")
			return
		}
		result, err = h.service.UploadVendorImage(ctx, file, fileHeader.Filename, entityID)

	case "user":
		if entityID == "" {
			response.SendError(w, r, http.StatusBadRequest, "MISSING_ENTITY_ID", "Entity ID is required for user uploads")
			return
		}
		result, err = h.service.UploadUserAvatar(ctx, file, fileHeader.Filename, entityID)

	default:
		// General upload
		opts := &UploadOptions{
			Folder:         folder,
			UniqueFilename: true,
			ResourceType:   "image",
			AllowedFormats: []string{"jpg", "jpeg", "png", "webp"},
			Transformation: "q_auto,f_auto",
		}
		result, err = h.service.UploadFile(ctx, file, fileHeader.Filename, opts)
	}

	if err != nil {
		response.SendError(w, r, http.StatusInternalServerError, "UPLOAD_FAILED", fmt.Sprintf("Failed to upload image: %v", err))
		return
	}

	// Return response
	uploadResponse := UploadImageResponse{
		PublicID:  result.PublicID,
		SecureURL: result.SecureURL,
		URL:       result.URL,
		Width:     result.Width,
		Height:    result.Height,
		Format:    result.Format,
	}

	response.SendSuccess(w, r, uploadResponse)
}

// DeleteImageRequest represents the delete image request
type DeleteImageRequest struct {
	PublicID string `json:"public_id"`
	URL      string `json:"url"` // Alternative: provide URL instead of public_id
}

// HandleDeleteImage handles image deletion requests
func (h *Handler) HandleDeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.SendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE method is allowed")
		return
	}

	var req DeleteImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	publicID := req.PublicID
	if publicID == "" && req.URL != "" {
		// Extract public ID from URL
		publicID = h.service.ExtractPublicIDFromURL(req.URL)
	}

	if publicID == "" {
		response.SendError(w, r, http.StatusBadRequest, "MISSING_PUBLIC_ID", "Public ID or URL is required")
		return
	}

	// Delete the image
	if err := h.service.DeleteFile(r.Context(), publicID); err != nil {
		response.SendError(w, r, http.StatusInternalServerError, "DELETE_FAILED", fmt.Sprintf("Failed to delete image: %v", err))
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Image deleted successfully"})
}

// GetTransformedURLRequest represents the get transformed URL request
type GetTransformedURLRequest struct {
	PublicID       string `json:"public_id"`
	Transformation string `json:"transformation,omitempty"` // e.g., "w_500,h_500,c_fill"
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	ThumbnailSize  int    `json:"thumbnail_size,omitempty"`
}

// GetTransformedURLResponse represents the transformed URL response
type GetTransformedURLResponse struct {
	URL string `json:"url"`
}

// HandleGetTransformedURL handles requests for generating transformed URLs
func (h *Handler) HandleGetTransformedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.SendError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	var req GetTransformedURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.PublicID == "" {
		response.SendError(w, r, http.StatusBadRequest, "MISSING_PUBLIC_ID", "Public ID is required")
		return
	}

	var url string
	var err error

	// Generate URL based on request parameters
	if req.ThumbnailSize > 0 {
		url, err = h.service.GetThumbnailURL(req.PublicID, req.ThumbnailSize)
	} else if req.Width > 0 && req.Height > 0 {
		url, err = h.service.GetOptimizedURL(req.PublicID, req.Width, req.Height)
	} else if req.Transformation != "" {
		url, err = h.service.GetTransformedURL(req.PublicID, req.Transformation)
	} else {
		url, err = h.service.GetTransformedURL(req.PublicID, "")
	}

	if err != nil {
		response.SendError(w, r, http.StatusInternalServerError, "URL_GENERATION_FAILED", fmt.Sprintf("Failed to generate URL: %v", err))
		return
	}

	response.SendSuccess(w, r, GetTransformedURLResponse{URL: url})
}
