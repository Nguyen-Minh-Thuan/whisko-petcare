package examples

import (
	"context"
	"log"
	"net/http"
	"time"

	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

// DemonstrateEnhancedErrorHandling shows how to use the enhanced error handling middleware
func DemonstrateEnhancedErrorHandling() {
	log.Println("Enhanced Error Handling Demo")
	log.Println("============================")

	// Create a sample handler that might panic
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Something went wrong!")
	})

	// Create a timeout handler
	timeoutHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second) // Simulate long operation
		w.Write([]byte("Should not reach here"))
	})

	// Create a validation error handler
	validationHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.HandleError(w, r, errors.NewValidationError("Name is required"))
	})

	// Create a custom error handler
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.HandleError(w, r, errors.NewUnauthorizedError("Please login first"))
	})

	// Create handlers that use ApiResponse for success cases
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"message": "Operation successful",
			"data":    []string{"item1", "item2", "item3"},
		}
		response.SendSuccess(w, r, data)
	})

	// Create pagination example handler
	paginationHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		}

		meta := &response.Meta{
			Page:       1,
			Limit:      10,
			Total:      2,
			TotalPages: 1,
		}

		response.SendSuccessWithMeta(w, r, users, meta)
	})

	// Create validation error example
	validationErrorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validationErrors := []response.ValidationError{
			{Field: "email", Message: "Email is required"},
			{Field: "name", Message: "Name must be at least 2 characters"},
		}
		response.SendValidationError(w, r, validationErrors)
	})

	// Create rate limited handler
	rateLimiter := middleware.NewRateLimiter(5, time.Minute)
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success!"))
	})

	// Setup middleware chain
	mux := http.NewServeMux()

	// Stack middleware in the correct order:
	// Request ID -> Rate Limiting -> Timeout -> Recovery -> Logging -> Your Handler
	chain := func(handler http.Handler) http.Handler {
		return middleware.RequestIDMiddleware(
			rateLimiter.Middleware(
				middleware.TimeoutMiddleware(5 * time.Second)(
					middleware.RecoveryMiddleware(
						middleware.LoggingMiddleware(handler),
					),
				),
			),
		)
	}

	// Register handlers
	mux.Handle("/panic", chain(panicHandler))
	mux.Handle("/timeout", chain(timeoutHandler))
	mux.Handle("/validation", chain(validationHandler))
	mux.Handle("/unauthorized", chain(customHandler))
	mux.Handle("/normal", chain(normalHandler))
	mux.Handle("/rate-limited", chain(normalHandler))
	mux.Handle("/success", chain(successHandler))
	mux.Handle("/pagination", chain(paginationHandler))
	mux.Handle("/validation-details", chain(validationErrorHandler))

	log.Println("Enhanced error handling demo endpoints available at:")
	log.Println("- /panic - Demonstrates panic recovery")
	log.Println("- /timeout - Demonstrates request timeout handling")
	log.Println("- /validation - Demonstrates validation errors")
	log.Println("- /unauthorized - Demonstrates authentication errors")
	log.Println("- /normal - Normal successful response")
	log.Println("- /rate-limited - Test rate limiting (5 requests per minute)")
	log.Println("- /success - ApiResponse success example")
	log.Println("- /pagination - ApiResponse with pagination metadata")
	log.Println("- /validation-details - Detailed validation errors")
}

// DatabaseOperationExample shows how to handle database errors
func DatabaseOperationExample(ctx context.Context) error {
	// Simulate a database operation that might fail
	err := simulateDBOperation()

	// Use the database error handler to convert to appropriate application error
	if err != nil {
		return middleware.DatabaseErrorHandler(err)
	}

	return nil
}

func simulateDBOperation() error {
	// This would be your actual database call
	// return db.Find(&users)

	// Simulate different types of database errors
	return &mockDBError{message: "connection timeout"}
}

type mockDBError struct {
	message string
}

func (e *mockDBError) Error() string {
	return e.message
}

// ErrorRecoveryExample shows how to gracefully handle different error scenarios
func ErrorRecoveryExample() {
	// Example of using different error types based on business logic

	// Validation error for missing required fields
	if err := validateUserInput("", "invalid-email"); err != nil {
		log.Printf("Validation failed: %v", err)
	}

	// Business logic error for duplicate resource
	if err := createUser("existing@email.com"); err != nil {
		log.Printf("User creation failed: %v", err)
	}

	// Authorization error for protected resources
	if err := accessProtectedResource("invalid-token"); err != nil {
		log.Printf("Access denied: %v", err)
	}
}

func validateUserInput(name, email string) error {
	if name == "" {
		return errors.NewValidationError("Name is required")
	}
	if email == "" || !isValidEmail(email) {
		return errors.NewValidationError("Valid email is required")
	}
	return nil
}

func createUser(email string) error {
	// Simulate checking for existing user
	if email == "existing@email.com" {
		return errors.NewConflictError("User with this email already exists")
	}
	return nil
}

func accessProtectedResource(token string) error {
	if token == "" {
		return errors.NewUnauthorizedError("Authentication token required")
	}
	if token == "invalid-token" {
		return errors.NewForbiddenError("Invalid or expired token")
	}
	return nil
}

func isValidEmail(email string) bool {
	return len(email) > 5 && email[0] != '@'
}
