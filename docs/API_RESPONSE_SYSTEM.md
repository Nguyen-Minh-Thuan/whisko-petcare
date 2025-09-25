# ApiResponse System Guide

This document explains the standardized ApiResponse system implemented for the Whisko Pet Care API.

## Overview

The ApiResponse system provides a consistent structure for all API responses, both success and error cases. This ensures a uniform experience for API consumers and makes error handling more predictable.

## Response Structure

### Success Response Format
```json
{
  "success": true,
  "data": { ... },
  "meta": { ... },
  "request_id": "a1b2c3d4",
  "timestamp": "2025-09-25T10:30:00.000Z"
}
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [...]
  },
  "request_id": "a1b2c3d4", 
  "timestamp": "2025-09-25T10:30:00.000Z"
}
```

## Response Components

### ApiResponse
- `success` (boolean) - Indicates if the operation was successful
- `data` (object, optional) - Contains the response data for successful operations
- `error` (object, optional) - Contains error information for failed operations
- `meta` (object, optional) - Contains metadata like pagination info
- `request_id` (string) - Unique identifier for request tracing
- `timestamp` (string) - ISO timestamp when the response was generated

### ApiError
- `code` (string) - Machine-readable error code
- `message` (string) - Human-readable error message
- `details` (object, optional) - Additional error details (e.g., validation errors)

### Meta (for pagination)
- `page` (number) - Current page number
- `limit` (number) - Number of items per page
- `total` (number) - Total number of items
- `total_pages` (number) - Total number of pages

## Usage in Controllers

### Import the Response Package
```go
import "whisko-petcare/pkg/response"
```

### Success Responses

#### Simple Success (200 OK)
```go
func (c *Controller) GetUser(w http.ResponseWriter, r *http.Request) {
    user := getUserFromDB()
    response.SendSuccess(w, r, user)
}
```

#### Created Response (201 Created)
```go
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
    userID := createUserInDB()
    data := map[string]interface{}{
        "id": userID,
        "message": "User created successfully",
    }
    response.SendCreated(w, r, data)
}
```

#### Success with Custom Status
```go
func (c *Controller) SomeOperation(w http.ResponseWriter, r *http.Request) {
    data := performOperation()
    response.SendSuccessWithStatus(w, r, http.StatusAccepted, data)
}
```

#### Success with Pagination
```go
func (c *Controller) ListUsers(w http.ResponseWriter, r *http.Request) {
    users := getUsersFromDB()
    meta := &response.Meta{
        Page:       1,
        Limit:      10,
        Total:      len(users),
        TotalPages: 1,
    }
    response.SendSuccessWithMeta(w, r, users, meta)
}
```

#### No Content Response (204)
```go
func (c *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
    updateUserInDB()
    response.SendNoContent(w, r)
}
```

### Error Responses

The error handling middleware automatically uses the ApiResponse format, so you continue to use:
```go
middleware.HandleError(w, r, errors.NewValidationError("Invalid input"))
```

#### Direct Error Responses (if needed)
```go
// Simple error
response.SendBadRequest(w, r, "Invalid request format")
response.SendUnauthorized(w, r, "Authentication required")
response.SendNotFound(w, r, "User not found")
response.SendConflict(w, r, "Email already exists")
response.SendInternalError(w, r, "Database connection failed")

// Validation errors with details
validationErrors := []response.ValidationError{
    {Field: "email", Message: "Email is required"},
    {Field: "name", Message: "Name must be at least 2 characters"},
}
response.SendValidationError(w, r, validationErrors)

// Custom error with details
response.SendErrorWithDetails(w, r, 422, "CUSTOM_ERROR", "Operation failed", customDetails)
```

## Example API Responses

### Successful User Creation
```json
{
  "success": true,
  "data": {
    "id": "user_1727261400000000000",
    "message": "User created successfully"
  },
  "request_id": "a1b2c3d4e5f6",
  "timestamp": "2025-09-25T10:30:00.123Z"
}
```

### User List with Pagination
```json
{
  "success": true,
  "data": [
    {
      "id": "user_1",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-09-25T09:00:00.000Z"
    },
    {
      "id": "user_2", 
      "name": "Jane Smith",
      "email": "jane@example.com",
      "created_at": "2025-09-25T09:15:00.000Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 2,
    "total_pages": 1
  },
  "request_id": "b2c3d4e5f6g7",
  "timestamp": "2025-09-25T10:30:00.456Z"
}
```

### Validation Error
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Email is required"
      },
      {
        "field": "name", 
        "message": "Name must be at least 2 characters"
      }
    ]
  },
  "request_id": "c3d4e5f6g7h8",
  "timestamp": "2025-09-25T10:30:00.789Z"
}
```

### Authentication Error
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  },
  "request_id": "d4e5f6g7h8i9",
  "timestamp": "2025-09-25T10:30:01.012Z"
}
```

### Server Error with Request Tracing
```json
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR", 
    "message": "Database connection failed"
  },
  "request_id": "e5f6g7h8i9j0",
  "timestamp": "2025-09-25T10:30:01.345Z"
}
```

## Benefits

### For Developers
- **Consistent Structure**: All responses follow the same format
- **Easy Debugging**: Request IDs enable request tracing across logs
- **Type Safety**: Strong typing for response components
- **Flexible**: Support for various response scenarios

### For API Consumers
- **Predictable**: Always know what to expect in responses
- **Rich Error Information**: Detailed error messages and validation feedback
- **Request Tracing**: Request IDs for support and debugging
- **Pagination Support**: Built-in metadata for paginated responses

### For Operations
- **Monitoring**: Consistent error codes for alerting
- **Debugging**: Request IDs correlate client requests with server logs
- **Analytics**: Structured data for response analysis

## Migration from Old Format

### Before (Direct JSON encoding)
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
    "id": user.ID,
    "name": user.Name,
})
```

### After (ApiResponse)
```go
userData := map[string]interface{}{
    "id": user.ID,
    "name": user.Name,
}
response.SendSuccess(w, r, userData)
```

## Best Practices

### 1. Use Appropriate Response Methods
- `SendSuccess()` for 200 OK responses
- `SendCreated()` for 201 Created responses  
- `SendNoContent()` for 204 No Content responses
- Use specific error methods like `SendNotFound()`, `SendValidationError()`

### 2. Include Meaningful Data
```go
// Good - descriptive response
data := map[string]interface{}{
    "user_id": userID,
    "message": "User created successfully",
    "next_steps": "Please verify your email address",
}
response.SendCreated(w, r, data)

// Avoid - minimal information
response.SendCreated(w, r, userID)
```

### 3. Use Metadata for Lists
```go
// Always include pagination info for list responses
meta := &response.Meta{
    Page:       page,
    Limit:      limit,
    Total:      totalCount,
    TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
}
response.SendSuccessWithMeta(w, r, items, meta)
```

### 4. Provide Detailed Validation Errors
```go
// Good - field-specific errors
validationErrors := []response.ValidationError{
    {Field: "email", Message: "Must be a valid email address"},
    {Field: "password", Message: "Must be at least 8 characters long"},
}
response.SendValidationError(w, r, validationErrors)

// Avoid - generic error
response.SendBadRequest(w, r, "Invalid input")
```

## Testing ApiResponse

You can test the ApiResponse format using curl:

```bash
# Success response
curl -X GET http://localhost:8080/users/123

# Error response  
curl -X GET http://localhost:8080/users/invalid-id

# Pagination response
curl -X GET http://localhost:8080/users

# Validation error
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{}'
```

The ApiResponse system provides a robust, consistent, and developer-friendly way to handle all API communications in your Whisko Pet Care application.