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
	"time"

	"github.com/google/uuid"
)

// PayoutConfig holds the configuration for PayOS payout integration
type PayoutConfig struct {
	ClientID    string
	APIKey      string
	ChecksumKey string
	BaseURL     string
	WebhookURL  string
}

// PayoutService handles real bank transfers via PayOS Payout API
type PayoutService struct {
	config     *PayoutConfig
	httpClient *http.Client
}

// CreatePayoutRequest represents the request to create a payout
type CreatePayoutRequest struct {
	ReferenceID     string   `json:"referenceId"`
	Amount          int      `json:"amount"`
	Description     string   `json:"description"`
	ToBin           string   `json:"toBin"`           // Bank code (e.g., "970415" for VietinBank)
	ToAccountNumber string   `json:"toAccountNumber"` // Recipient bank account number
	Category        []string `json:"category,omitempty"`
}

// PayoutResponse represents the response from PayOS payout API
type PayoutResponse struct {
	Code string          `json:"code"`
	Desc string          `json:"desc"`
	Data PayoutDataModel `json:"data"`
}

// PayoutDataModel represents the payout data in the response
type PayoutDataModel struct {
	ID            string                `json:"id"`
	ReferenceID   string                `json:"referenceId"`
	Transactions  []PayoutTransaction   `json:"transactions"`
	Category      []string              `json:"category"`
	ApprovalState string                `json:"approvalState"`
	CreatedAt     time.Time             `json:"createdAt"`
}

// PayoutTransaction represents a single transaction in the payout
type PayoutTransaction struct {
	ID                  string    `json:"id"`
	ReferenceID         string    `json:"referenceId"`
	Amount              int       `json:"amount"`
	Description         string    `json:"description"`
	ToBin               string    `json:"toBin"`
	ToAccountNumber     string    `json:"toAccountNumber"`
	ToAccountName       string    `json:"toAccountName"`
	Reference           string    `json:"reference"`           // Transaction reference from bank
	TransactionDatetime time.Time `json:"transactionDatetime"` // When transaction was processed
	ErrorMessage        string    `json:"errorMessage,omitempty"`
	ErrorCode           string    `json:"errorCode,omitempty"`
	State               string    `json:"state"` // SUCCEEDED, FAILED, PROCESSING
}

// PayoutInfo represents detailed information about a payout
type PayoutInfo struct {
	PayoutID      string
	ReferenceID   string
	TransferID    string // Bank transaction reference
	Amount        int
	Status        string // SUCCEEDED, FAILED, PROCESSING
	ErrorMessage  string
	ProcessedAt   time.Time
}

// BankCodeMap maps common Vietnamese bank names to PayOS bank codes (BIN)
var BankCodeMap = map[string]string{
	// Major banks
	"VietinBank":       "970415",
	"Vietcombank":      "970436",
	"BIDV":             "970418",
	"Agribank":         "970405",
	"Techcombank":      "970407",
	"MB Bank":          "970422",
	"ACB":              "970416",
	"VPBank":           "970432",
	"Sacombank":        "970403",
	"VIB":              "970441",
	"HDBank":           "970437",
	"TPBank":           "970423",
	"SHB":              "970443",
	"SeABank":          "970440",
	"OCB":              "970448",
	"MSB":              "970426",
	"VietCapitalBank":  "970454",
	"SCB":              "970429",
	"LienVietPostBank": "970449",
	"VietABank":        "970427",
	"ABBank":           "970425",
	"NCB":              "970419",
	"BacABank":         "970409",
	"PVcomBank":        "970412",
	"Eximbank":         "970431",
	"KienlongBank":     "970452",
	"GPBank":           "970408",
	"PGBank":           "970430",
	"BaoVietBank":      "970438",
	"CAKE":             "546034",
	"Ubank":            "546035",
	"Timo":             "963388",
	"ViettelMoney":     "971005",
}

// GetBankCode returns the PayOS bank code (BIN) for a given bank name
// Returns empty string if bank is not found
func GetBankCode(bankName string) string {
	return BankCodeMap[bankName]
}

// NewPayoutService creates a new payout service instance
func NewPayoutService(config *PayoutConfig) *PayoutService {
	if config.BaseURL == "" {
		config.BaseURL = "https://api-merchant.payos.vn"
	}

	return &PayoutService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessPayout executes a real bank transfer to the vendor's account
func (s *PayoutService) ProcessPayout(ctx context.Context, payoutID, vendorBankName, vendorAccountNumber, vendorAccountName string, amount int, description string) (*PayoutInfo, error) {
	// Get bank code from bank name
	bankCode := GetBankCode(vendorBankName)
	if bankCode == "" {
		return nil, fmt.Errorf("unsupported bank: %s. Please use one of the supported Vietnamese banks", vendorBankName)
	}

	// Validate amount (PayOS limits)
	if amount <= 0 {
		return nil, fmt.Errorf("payout amount must be greater than 0")
	}
	if amount > 500000000 { // Max 500M VND
		return nil, fmt.Errorf("payout amount must not exceed 500,000,000 VND")
	}

	// Create payout request
	payoutReq := CreatePayoutRequest{
		ReferenceID:     payoutID,
		Amount:          amount,
		Description:     description,
		ToBin:           bankCode,
		ToAccountNumber: vendorAccountNumber,
		Category:        []string{"vendor_payout"},
	}

	// Call PayOS API
	response, err := s.createPayout(ctx, &payoutReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payout: %w", err)
	}

	// Parse response
	info := &PayoutInfo{
		PayoutID:    payoutID,
		ReferenceID: response.Data.ReferenceID,
		Amount:      amount,
	}

	if len(response.Data.Transactions) > 0 {
		transaction := response.Data.Transactions[0]
		info.TransferID = transaction.Reference
		info.Status = transaction.State
		info.ErrorMessage = transaction.ErrorMessage
		info.ProcessedAt = transaction.TransactionDatetime
	}

	// Map approval state to status
	if response.Data.ApprovalState == "APPROVED" && info.Status == "" {
		info.Status = "PROCESSING"
	} else if info.Status == "" {
		info.Status = "PENDING"
	}

	return info, nil
}

// createPayout makes the HTTP request to PayOS payout API
func (s *PayoutService) createPayout(ctx context.Context, req *CreatePayoutRequest) (*PayoutResponse, error) {
	// Marshal request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/payouts", s.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate idempotency key (unique per request)
	idempotencyKey := uuid.New().String()

	// Generate signature for request authentication
	signature := s.generateSignature(req)

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-client-id", s.config.ClientID)
	httpReq.Header.Set("x-api-key", s.config.APIKey)
	httpReq.Header.Set("x-idempotency-key", idempotencyKey)
	httpReq.Header.Set("x-signature", signature)

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PayOS API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var payoutResp PayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check response code
	if payoutResp.Code != "00" {
		return nil, fmt.Errorf("PayOS payout failed: %s - %s", payoutResp.Code, payoutResp.Desc)
	}

	return &payoutResp, nil
}

// generateSignature creates HMAC-SHA256 signature for PayOS authentication
func (s *PayoutService) generateSignature(req *CreatePayoutRequest) string {
	// Create data string for signature
	// Format: referenceId=XXX&amount=XXX&description=XXX&toBin=XXX&toAccountNumber=XXX
	data := fmt.Sprintf("amount=%d&description=%s&referenceId=%s&toAccountNumber=%s&toBin=%s",
		req.Amount,
		req.Description,
		req.ReferenceID,
		req.ToAccountNumber,
		req.ToBin,
	)

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(s.config.ChecksumKey))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// GetPayoutInfo retrieves the status of a payout from PayOS
func (s *PayoutService) GetPayoutInfo(ctx context.Context, payoutID string) (*PayoutInfo, error) {
	// Create HTTP request
	url := fmt.Sprintf("%s/v1/payouts/%s", s.config.BaseURL, payoutID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("x-client-id", s.config.ClientID)
	httpReq.Header.Set("x-api-key", s.config.APIKey)

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PayOS API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var payoutResp PayoutResponse
	if err := json.Unmarshal(respBody, &payoutResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Parse into PayoutInfo
	info := &PayoutInfo{
		PayoutID:    payoutID,
		ReferenceID: payoutResp.Data.ReferenceID,
	}

	if len(payoutResp.Data.Transactions) > 0 {
		transaction := payoutResp.Data.Transactions[0]
		info.TransferID = transaction.Reference
		info.Amount = transaction.Amount
		info.Status = transaction.State
		info.ErrorMessage = transaction.ErrorMessage
		info.ProcessedAt = transaction.TransactionDatetime
	}

	return info, nil
}
