package errors

import "fmt"

// ApplicationError represents a domain-specific error
type ApplicationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *ApplicationError) Error() string {
	return e.Message
}

// Error constructors
func NewValidationError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Status:  400,
	}
}

func NewNotFoundError(resource string) *ApplicationError {
	return &ApplicationError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Status:  404,
	}
}

func NewConflictError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "CONFLICT",
		Message: message,
		Status:  409,
	}
}

func NewInternalError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  500,
	}
}
