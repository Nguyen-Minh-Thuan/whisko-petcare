# Enhanced Error Handling Guide

This document explains the enhanced error handling features added to the Whisko Pet Care API.

## Overview

The error handling system has been significantly enhanced with the following features:

- **Request ID tracking** - Each request gets a unique ID for better tracing
- **Panic recovery** - Automatic recovery from panics with full stack traces
- **Request timeout handling** - Configurable timeouts with proper error responses
- **Rate limiting** - Protection against abuse with configurable limits
- **Enhanced logging** - Structured logs with request context
- **Multiple error types** - Support for all common HTTP error scenarios
- **Database error mapping** - Automatic conversion of database errors to application errors

## New Error Types

The following error types have been added to `pkg/errors/errors.go`:

```go
// Authentication & Authorization
NewUnauthorizedError(message)      // 401 - Authentication required
NewForbiddenError(message)         // 403 - Access denied

// Request Issues
NewBadRequestError(message)        // 400 - Generic bad request
NewRequestTimeoutError(message)    // 408 - Request timeout
NewMethodNotAllowedError(message)  // 405 - HTTP method not allowed
NewUnprocessableEntityError(message) // 422 - Validation failed
NewTooManyRequestsError(message)   // 429 - Rate limit exceeded

// Server Issues
NewServiceUnavailableError(message) // 503 - Service temporarily unavailable
```

## Enhanced Middleware

### 1. Request ID Middleware
Generates unique request IDs for tracking requests across logs:

```go
// Usage in main.go
handler := middleware.RequestIDMiddleware(yourHandler)
```

### 2. Recovery Middleware
Enhanced panic recovery with better logging:

```go
// Usage in main.go  
handler := middleware.RecoveryMiddleware(yourHandler)
```

### 3. Rate Limiting Middleware
Configurable rate limiting per IP address:

```go
// Usage in main.go
rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute
handler := rateLimiter.Middleware(yourHandler)
```

### 4. Timeout Middleware
Request timeout handling:

```go
// Usage in main.go
handler := middleware.TimeoutMiddleware(30 * time.Second)(yourHandler)
```

### 5. Enhanced Logging Middleware
Logs with request IDs and timing information:

```go
// Usage in main.go
handler := middleware.LoggingMiddleware(yourHandler)
```

## Middleware Chain Setup

Recommended order for stacking middleware:

```go
// In your main.go, create a middleware chain function:
func middlewareChain(handler http.Handler) http.Handler {
    return middleware.RequestIDMiddleware(
        middleware.NewRateLimiter(100, time.Minute).Middleware(
            middleware.TimeoutMiddleware(30 * time.Second)(
                middleware.RecoveryMiddleware(
                    middleware.LoggingMiddleware(handler),
                ),
            ),
        ),
    )
}

// Apply to routes:
mux.Handle("/users", middlewareChain(yourUserHandler))
```

## Enhanced HandleError Function

The `HandleError` function now requires the request parameter and provides enhanced logging:

```go
// OLD usage:
middleware.HandleError(w, err)

// NEW usage:
middleware.HandleError(w, r, err)
```

**Features:**
- Includes request ID in response
- Logs request method and path
- Provides detailed error context

## Database Error Handling

Use `DatabaseErrorHandler` to convert database errors:

```go
func (repo *UserRepository) GetUser(id string) (*User, error) {
    user, err := repo.db.FindByID(id)
    if err != nil {
        return nil, middleware.DatabaseErrorHandler(err)
    }
    return user, nil
}
```

This automatically maps database errors to appropriate HTTP errors:
- Connection errors → 503 Service Unavailable
- Timeout errors → 408 Request Timeout  
- Duplicate key errors → 409 Conflict
- Not found errors → 404 Not Found

## Example Usage Patterns

### 1. Authentication Error
```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            middleware.HandleError(w, r, errors.NewUnauthorizedError("Authentication required"))
            return
        }
        // ... validate token
        next.ServeHTTP(w, r)
    })
}
```

### 2. Validation Error
```go
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
        return
    }
    
    if user.Email == "" {
        middleware.HandleError(w, r, errors.NewValidationError("Email is required"))
        return
    }
    // ... continue processing
}
```

### 3. Business Logic Error
```go
func (s *UserService) CreateUser(ctx context.Context, cmd CreateUser) error {
    // Check if user already exists
    existing, err := s.repo.FindByEmail(cmd.Email)
    if err != nil {
        return middleware.DatabaseErrorHandler(err)
    }
    if existing != nil {
        return errors.NewConflictError("User with this email already exists")
    }
    // ... create user
    return nil
}
```

## Log Output Examples

With the enhanced error handling, you'll see logs like:

```
[a1b2c3d4] Started POST /users from 127.0.0.1:54321
[a1b2c3d4] Error 400: Email is required (Code: VALIDATION_ERROR) for POST /users  
[a1b2c3d4] Completed POST /users in 15.2ms

[e5f6g7h8] Started GET /users/invalid-id from 127.0.0.1:54322
[e5f6g7h8] Error 404: user not found (Code: NOT_FOUND) for GET /users/invalid-id
[e5f6g7h8] Completed GET /users/invalid-id in 8.7ms

[i9j0k1l2] PANIC: runtime error: invalid memory address
Stack Trace:
goroutine 1 [running]:
...
```

## Testing Error Handling

You can test the error handling with these endpoints (add to your main.go for development):

```go
// Add these demo endpoints for testing:
mux.HandleFunc("/demo/panic", func(w http.ResponseWriter, r *http.Request) {
    panic("Intentional panic for testing")
})

mux.HandleFunc("/demo/validation", func(w http.ResponseWriter, r *http.Request) {
    middleware.HandleError(w, r, errors.NewValidationError("Demo validation error"))
})

mux.HandleFunc("/demo/unauthorized", func(w http.ResponseWriter, r *http.Request) {
    middleware.HandleError(w, r, errors.NewUnauthorizedError("Demo auth error"))
})

mux.HandleFunc("/demo/timeout", func(w http.ResponseWriter, r *http.Request) {
    time.Sleep(35 * time.Second) // Will trigger timeout middleware
    w.Write([]byte("Should not reach here"))
})
```

## Configuration Recommendations

For production use:

```go
// Rate limiting: 1000 requests per minute per IP
rateLimiter := middleware.NewRateLimiter(1000, time.Minute)

// Request timeout: 30 seconds
timeoutMiddleware := middleware.TimeoutMiddleware(30 * time.Second)

// Server timeouts
server := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,  
    IdleTimeout:  120 * time.Second,
}
```

For development:
```go
// More lenient rate limiting: 100 requests per minute
rateLimiter := middleware.NewRateLimiter(100, time.Minute)

// Shorter timeout for faster feedback: 10 seconds  
timeoutMiddleware := middleware.TimeoutMiddleware(10 * time.Second)
```

## Migration Steps

To migrate your existing code:

1. **Update HandleError calls**: Add the request parameter to all `middleware.HandleError(w, err)` calls → `middleware.HandleError(w, r, err)`

2. **Add middleware chain**: Create a middleware chain function and apply it to your routes

3. **Update error types**: Replace basic errors with more specific error types where appropriate

4. **Add database error handling**: Wrap database operations with `DatabaseErrorHandler`

5. **Test thoroughly**: Use the demo endpoints to verify all error scenarios work correctly

The enhanced error handling provides better debugging capabilities, improved user experience, and production-ready error management for your API.