package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TestApiResponseSystem demonstrates the new ApiResponse format
func TestApiResponseSystem() {
	baseURL := "http://localhost:8080"

	fmt.Println("ApiResponse System Test")
	fmt.Println("=======================")

	// Test 1: Successful user creation
	fmt.Println("\n1. Testing User Creation (ApiResponse format):")
	testUserCreation(baseURL)

	// Test 2: Get user (success response)
	fmt.Println("\n2. Testing Get User (success response):")
	userID := testUserCreationAndGetID(baseURL)
	if userID != "" {
		testGetUser(baseURL, userID)
	}

	// Test 3: List users (pagination metadata)
	fmt.Println("\n3. Testing List Users (with pagination metadata):")
	testListUsers(baseURL)

	// Test 4: Validation error (detailed error response)
	fmt.Println("\n4. Testing Validation Error (detailed response):")
	testApiValidationError(baseURL)

	// Test 5: Not found error
	fmt.Println("\n5. Testing Not Found Error:")
	testNotFoundError(baseURL)

	// Test 6: Update user (success response)
	fmt.Println("\n6. Testing Update User (success response):")
	if userID != "" {
		testUpdateUser(baseURL, userID)
	}
}

func testUserCreation(baseURL string) {
	payload := map[string]interface{}{
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"phone":   "+1234567890",
		"address": "123 Main St, Anytown, USA",
	}

	resp, body := makeAPIRequest("POST", baseURL+"/users", payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	// Parse and validate ApiResponse structure
	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "success")
	}
}

func testUserCreationAndGetID(baseURL string) string {
	payload := map[string]interface{}{
		"name":  "Test User",
		"email": "test.user@example.com",
	}

	resp, body := makeAPIRequest("POST", baseURL+"/users", payload)
	if resp.StatusCode == 201 {
		var apiResp map[string]interface{}
		if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
			if data, ok := apiResp["data"].(map[string]interface{}); ok {
				if userID, ok := data["id"].(string); ok {
					return userID
				}
			}
		}
	}
	return ""
}

func testGetUser(baseURL, userID string) {
	resp, body := makeAPIRequest("GET", baseURL+"/users/"+userID, nil)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "success")
	}
}

func testListUsers(baseURL string) {
	resp, body := makeAPIRequest("GET", baseURL+"/users", nil)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "success")
		// Check for pagination metadata
		if meta, exists := apiResp["meta"]; exists {
			fmt.Println("✅ Pagination metadata found:", meta)
		}
	}
}

func testApiValidationError(baseURL string) {
	// Send invalid request (empty JSON)
	payload := map[string]interface{}{}

	resp, body := makeAPIRequest("POST", baseURL+"/users", payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "error")
	}
}

func testNotFoundError(baseURL string) {
	resp, body := makeAPIRequest("GET", baseURL+"/users/nonexistent-id", nil)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "error")
	}
}

func testUpdateUser(baseURL, userID string) {
	payload := map[string]interface{}{
		"name":  "John Updated",
		"phone": "+1987654321",
	}

	resp, body := makeAPIRequest("PUT", baseURL+"/users/"+userID, payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body:\n%s\n", formatJSON(body))

	var apiResp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &apiResp); err == nil {
		validateApiResponseStructure(apiResp, "success")
	}
}

func makeAPIRequest(method, url string, payload interface{}) (*http.Response, string) {
	var body io.Reader

	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, ""
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return nil, ""
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	return resp, string(responseBody)
}

func validateApiResponseStructure(apiResp map[string]interface{}, expectedType string) {
	fmt.Println("\n🔍 Validating ApiResponse structure:")

	// Check required fields
	if success, exists := apiResp["success"]; exists {
		fmt.Printf("✅ success field: %v\n", success)
	} else {
		fmt.Println("❌ Missing 'success' field")
	}

	if requestID, exists := apiResp["request_id"]; exists {
		fmt.Printf("✅ request_id field: %s\n", requestID)
	} else {
		fmt.Println("❌ Missing 'request_id' field")
	}

	if timestamp, exists := apiResp["timestamp"]; exists {
		fmt.Printf("✅ timestamp field: %s\n", timestamp)
	} else {
		fmt.Println("❌ Missing 'timestamp' field")
	}

	// Check type-specific fields
	if expectedType == "success" {
		if data, exists := apiResp["data"]; exists {
			fmt.Printf("✅ data field present: %T\n", data)
		} else {
			fmt.Println("❌ Missing 'data' field for success response")
		}

		if meta, exists := apiResp["meta"]; exists {
			fmt.Printf("✅ meta field present: %T\n", meta)
		}
	} else if expectedType == "error" {
		if errorData, exists := apiResp["error"]; exists {
			fmt.Printf("✅ error field present: %T\n", errorData)

			// Validate error structure
			if errorObj, ok := errorData.(map[string]interface{}); ok {
				if code, exists := errorObj["code"]; exists {
					fmt.Printf("  ✅ error.code: %s\n", code)
				}
				if message, exists := errorObj["message"]; exists {
					fmt.Printf("  ✅ error.message: %s\n", message)
				}
			}
		} else {
			fmt.Println("❌ Missing 'error' field for error response")
		}
	}

	fmt.Println("🏁 Validation complete")
}

func formatJSON(jsonStr string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsonStr // Return as-is if not valid JSON
	}

	formatted, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return jsonStr
	}

	return string(formatted)
}

// ShowApiResponseExamples displays example API responses
func ShowApiResponseExamples() {
	fmt.Println("ApiResponse Examples")
	fmt.Println("===================")

	fmt.Println("\n📋 Success Response Example:")
	successExample := `{
  "success": true,
  "data": {
    "id": "user_1727261400000000000",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-09-25T10:30:00.000Z"
  },
  "request_id": "a1b2c3d4e5f6",
  "timestamp": "2025-09-25T10:30:00.123Z"
}`
	fmt.Println(successExample)

	fmt.Println("\n📄 Paginated Response Example:")
	paginationExample := `{
  "success": true,
  "data": [
    {"id": "user_1", "name": "John", "email": "john@example.com"},
    {"id": "user_2", "name": "Jane", "email": "jane@example.com"}
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 2,
    "total_pages": 1
  },
  "request_id": "b2c3d4e5f6g7",
  "timestamp": "2025-09-25T10:30:00.456Z"
}`
	fmt.Println(paginationExample)

	fmt.Println("\n❌ Error Response Example:")
	errorExample := `{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {"field": "email", "message": "Email is required"},
      {"field": "name", "message": "Name must be at least 2 characters"}
    ]
  },
  "request_id": "c3d4e5f6g7h8", 
  "timestamp": "2025-09-25T10:30:00.789Z"
}`
	fmt.Println(errorExample)
}
