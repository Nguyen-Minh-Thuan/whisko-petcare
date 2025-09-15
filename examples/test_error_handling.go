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

// TestErrorHandling demonstrates different error scenarios in the API
func TestErrorHandling() {
	baseURL := "http://localhost:8080"

	fmt.Println("Testing Error Handling in CQRS+ES API")
	fmt.Println("==========================================")

	// Test 1: Validation Error - Empty name
	fmt.Println("\n1. Testing Validation Error (empty name):")
	testValidationError(baseURL)

	// Test 2: Validation Error - Invalid JSON
	fmt.Println("\n2. Testing Invalid JSON:")
	testInvalidJSON(baseURL)

	// Test 3: Not Found Error
	fmt.Println("\n3. Testing Not Found Error:")
	testNotFound(baseURL)

	// Test 4: Successful request for comparison
	fmt.Println("\n4. Testing Successful Create:")
	userID := testSuccessfulCreate(baseURL)

	// Test 5: Update Error
	fmt.Println("\n5. Testing Update Error:")
	testUpdateError(baseURL, userID)
}

func testValidationError(baseURL string) {
	payload := map[string]interface{}{
		"name":  "", // Empty name should trigger validation error
		"email": "test@example.com",
	}

	resp, body := makeRequest("POST", baseURL+"/api/v1/users", payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", body)
}

func testInvalidJSON(baseURL string) {
	// Send malformed JSON
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
}

func testNotFound(baseURL string) {
	resp, body := makeRequest("GET", baseURL+"/api/v1/users/non-existent-id", nil)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", body)
}

func testSuccessfulCreate(baseURL string) string {
	payload := map[string]interface{}{
		"name":  "John Cena",
		"email": "ucancme@example.com",
	}

	resp, body := makeRequest("POST", baseURL+"/api/v1/users", payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	if resp.StatusCode == 201 {
		var user map[string]interface{}
		json.Unmarshal([]byte(body), &user)
		userID := user["id"].(string)
		fmt.Printf("Created user ID: %s\n", userID)
		return userID
	}

	fmt.Printf("Response: %s\n", body)
	return ""
}

func testUpdateError(baseURL, userID string) {
	if userID == "" {
		fmt.Println("Skipping update test - no user ID")
		return
	}

	payload := map[string]interface{}{
		"name":  "", // Empty name should trigger validation error
		"email": "updated@example.com",
	}

	resp, body := makeRequest("PUT", baseURL+"/api/v1/users/"+userID, payload)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", body)
}

func makeRequest(method, url string, payload interface{}) (*http.Response, string) {
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
