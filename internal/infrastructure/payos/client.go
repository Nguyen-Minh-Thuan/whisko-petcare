package payos

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PayOSConfig holds the configuration for PayOS integration
type PayOSConfig struct {
	ClientID     string
	APIKey       string
	ChecksumKey  string
	PartnerCode  string
	BaseURL      string
	WebhookURL   string
	ReturnURL    string
	CancelURL    string
}

// PayOSClient is the HTTP client for PayOS API
type PayOSClient struct {
	config     *PayOSConfig
	httpClient *http.Client
}

// PaymentItem represents an item in the payment request
type PaymentItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"` // Amount in VND
}

// CreatePaymentRequest represents the PayOS payment creation request
type CreatePaymentRequest struct {
	OrderCode   int64         `json:"orderCode"`
	Amount      int           `json:"amount"`      // Total amount in VND
	Description string        `json:"description"`
	Items       []PaymentItem `json:"items"`
	ReturnURL   string        `json:"returnUrl"`
	CancelURL   string        `json:"cancelUrl"`
}

// CreatePaymentResponse represents the PayOS payment creation response
type CreatePaymentResponse struct {
	Code    string      `json:"code"`
	Desc    string      `json:"desc"`
	Data    PaymentData `json:"data"`
	Success bool        `json:"success"`
}

// PaymentData represents the payment data in the response
type PaymentData struct {
	Bin           string `json:"bin"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	Amount        int    `json:"amount"`
	Description   string `json:"description"`
	OrderCode     int64  `json:"orderCode"`
	Currency      string `json:"currency"`
	PaymentLinkId string `json:"paymentLinkId"`
	Status        string `json:"status"`
	CheckoutUrl   string `json:"checkoutUrl"`
	QrCode        string `json:"qrCode"`
}

// PaymentInfoResponse represents the PayOS payment info response
type PaymentInfoResponse struct {
	Code    string            `json:"code"`
	Desc    string            `json:"desc"`
	Data    PaymentInfoData   `json:"data"`
	Success bool              `json:"success"`
}

// PaymentInfoData represents the payment info data
type PaymentInfoData struct {
	OrderCode     int64                  `json:"orderCode"`
	Amount        int                    `json:"amount"`
	AmountPaid    int                    `json:"amountPaid"`
	AmountRemaining int                  `json:"amountRemaining"`
	Status        string                 `json:"status"`
	CreatedAt     string                 `json:"createdAt"`
	Transactions  []PaymentTransaction   `json:"transactions"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	Reference       string `json:"reference"`
	Amount          int    `json:"amount"`
	AccountNumber   string `json:"accountNumber"`
	Description     string `json:"description"`
	TransactionDateTime string `json:"transactionDateTime"`
}

// WebhookData represents the webhook payload from PayOS
type WebhookData struct {
	OrderCode   int64  `json:"orderCode"`
	Amount      int    `json:"amount"`
	Description string `json:"description"`
	AccountNumber string `json:"accountNumber"`
	Reference   string `json:"reference"`
	TransactionDateTime string `json:"transactionDateTime"`
	Currency    string `json:"currency"`
	PaymentLinkId string `json:"paymentLinkId"`
	Code        string `json:"code"`
	Desc        string `json:"desc"`
	Success     bool   `json:"success"`
}

// NewPayOSClient creates a new PayOS client
func NewPayOSClient(config *PayOSConfig) *PayOSClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api-merchant.payos.vn"
	}

	return &PayOSClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePayment creates a new payment link in PayOS
func (c *PayOSClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Set default URLs if not provided
	if req.ReturnURL == "" {
		req.ReturnURL = c.config.ReturnURL
	}
	if req.CancelURL == "" {
		req.CancelURL = c.config.CancelURL
	}

	// Create the request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/v2/payment-requests", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate signature
	signature := c.generateSignature("POST", "/v2/payment-requests", string(reqBody))

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-client-id", c.config.ClientID)
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("x-partner-code", c.config.PartnerCode)
	httpReq.Header.Set("x-signature", signature)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var paymentResp CreatePaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, paymentResp.Desc)
	}

	return &paymentResp, nil
}

// GetPaymentInfo retrieves payment information from PayOS
func (c *PayOSClient) GetPaymentInfo(ctx context.Context, orderCode int64) (*PaymentInfoResponse, error) {
	url := fmt.Sprintf("%s/v2/payment-requests/%d", c.config.BaseURL, orderCode)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate signature for GET request
	signature := c.generateSignature("GET", fmt.Sprintf("/v2/payment-requests/%d", orderCode), "")

	// Set headers
	httpReq.Header.Set("x-client-id", c.config.ClientID)
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("x-partner-code", c.config.PartnerCode)
	httpReq.Header.Set("x-signature", signature)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var infoResp PaymentInfoResponse
	if err := json.Unmarshal(respBody, &infoResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, infoResp.Desc)
	}

	return &infoResp, nil
}

// CancelPayment cancels a payment in PayOS
func (c *PayOSClient) CancelPayment(ctx context.Context, orderCode int64, cancelReason string) error {
	url := fmt.Sprintf("%s/v2/payment-requests/%d/cancel", c.config.BaseURL, orderCode)

	cancelReq := map[string]string{
		"cancellationReason": cancelReason,
	}

	reqBody, err := json.Marshal(cancelReq)
	if err != nil {
		return fmt.Errorf("failed to marshal cancel request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Generate signature
	signature := c.generateSignature("POST", fmt.Sprintf("/v2/payment-requests/%d/cancel", orderCode), string(reqBody))

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-client-id", c.config.ClientID)
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("x-partner-code", c.config.PartnerCode)
	httpReq.Header.Set("x-signature", signature)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cancel request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// VerifyWebhookSignature verifies the webhook signature from PayOS
func (c *PayOSClient) VerifyWebhookSignature(webhookBody []byte, signature string) bool {
	expectedSignature := c.generateWebhookSignature(webhookBody)
	return expectedSignature == signature
}

// generateSignature generates HMAC SHA256 signature for API requests
func (c *PayOSClient) generateSignature(method, path, body string) string {
	// Create data string: method + path + timestamp + body
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	data := method + path + timestamp + body

	// Create HMAC SHA256 hash
	h := hmac.New(sha256.New, []byte(c.config.ChecksumKey))
	h.Write([]byte(data))
	
	return hex.EncodeToString(h.Sum(nil))
}

// generateWebhookSignature generates signature for webhook verification
func (c *PayOSClient) generateWebhookSignature(body []byte) string {
	h := hmac.New(sha256.New, []byte(c.config.ChecksumKey))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// sortDataForSignature sorts data keys for consistent signature generation
func (c *PayOSClient) sortDataForSignature(data map[string]interface{}) string {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, data[k]))
	}

	return strings.Join(parts, "&")
}