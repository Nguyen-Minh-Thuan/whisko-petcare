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

// NewUnauthorizedError creates a 401 Unauthorized error
func NewUnauthorizedError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Status:  401,
	}
}

// NewForbiddenError creates a 403 Forbidden error
func NewForbiddenError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "FORBIDDEN",
		Message: message,
		Status:  403,
	}
}

// NewUnprocessableEntityError creates a 422 Unprocessable Entity error
func NewUnprocessableEntityError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "UNPROCESSABLE_ENTITY",
		Message: message,
		Status:  422,
	}
}

// NewTooManyRequestsError creates a 429 Too Many Requests error
func NewTooManyRequestsError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "TOO_MANY_REQUESTS",
		Message: message,
		Status:  429,
	}
}

// NewRequestTimeoutError creates a 408 Request Timeout error
func NewRequestTimeoutError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "REQUEST_TIMEOUT",
		Message: message,
		Status:  408,
	}
}

// NewBadRequestError creates a 400 Bad Request error (more specific than validation error)
func NewBadRequestError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "BAD_REQUEST",
		Message: message,
		Status:  400,
	}
}

// NewServiceUnavailableError creates a 503 Service Unavailable error
func NewServiceUnavailableError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: message,
		Status:  503,
	}
}

// NewMethodNotAllowedError creates a 405 Method Not Allowed error
func NewMethodNotAllowedError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    "METHOD_NOT_ALLOWED",
		Message: message,
		Status:  405,
	}
}
