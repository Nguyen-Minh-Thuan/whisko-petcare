package response

import (
	"encoding/json"
	"net/http"
	"time"

	"whisko-petcare/pkg/middleware"
)

// ApiResponse represents a standardized API response structure
type ApiResponse struct {
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Error     *ApiError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ApiError represents error details in the API response
type ApiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta contains metadata about the response
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SendSuccess sends a successful API response
func SendSuccess(w http.ResponseWriter, r *http.Request, data interface{}) {
	SendSuccessWithStatus(w, r, http.StatusOK, data)
}

// SendSuccessWithStatus sends a successful API response with custom status code
func SendSuccessWithStatus(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	response := ApiResponse{
		RequestID: middleware.GetRequestID(r.Context()),
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SendSuccessWithMeta sends a successful API response with metadata (for pagination)
func SendSuccessWithMeta(w http.ResponseWriter, r *http.Request, data interface{}, meta *Meta) {
	response := ApiResponse{
		RequestID: middleware.GetRequestID(r.Context()),
		Success:   true,
		Meta:      meta,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendCreated sends a 201 Created response
func SendCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	SendSuccessWithStatus(w, r, http.StatusCreated, data)
}

// SendNoContent sends a 204 No Content response
func SendNoContent(w http.ResponseWriter, r *http.Request) {
	response := ApiResponse{
		RequestID: middleware.GetRequestID(r.Context()),
		Success:   true,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(response)
}

// SendError sends an error API response
func SendError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string) {
	SendErrorWithDetails(w, r, statusCode, code, message, nil)
}

// SendErrorWithDetails sends an error API response with additional details
func SendErrorWithDetails(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, details interface{}) {
	response := ApiResponse{
		RequestID: middleware.GetRequestID(r.Context()),
		Success:   false,
		Error: &ApiError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SendValidationError sends a validation error response
func SendValidationError(w http.ResponseWriter, r *http.Request, errors []ValidationError) {
	SendErrorWithDetails(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}

// Common response builders for frequently used responses

// SendBadRequest sends a 400 Bad Request response
func SendBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusBadRequest, "BAD_REQUEST", message)
}

// SendUnauthorized sends a 401 Unauthorized response
func SendUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// SendForbidden sends a 403 Forbidden response
func SendForbidden(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusForbidden, "FORBIDDEN", message)
}

// SendNotFound sends a 404 Not Found response
func SendNotFound(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusNotFound, "NOT_FOUND", message)
}

// SendConflict sends a 409 Conflict response
func SendConflict(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusConflict, "CONFLICT", message)
}

// SendInternalError sends a 500 Internal Server Error response
func SendInternalError(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// SendServiceUnavailable sends a 503 Service Unavailable response
func SendServiceUnavailable(w http.ResponseWriter, r *http.Request, message string) {
	SendError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
