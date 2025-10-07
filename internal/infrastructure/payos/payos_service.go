package payos

import (
	"context"
	"fmt"
	"strconv"

	"whisko-petcare/internal/domain/aggregate"

	payossdk "github.com/payOSHQ/payos-lib-golang"
)

// Service wraps the official PayOS SDK
type Service struct {
	initialized bool
	config      *Config
}

// Config holds the configuration for PayOS integration
type Config struct {
	ClientID    string
	APIKey      string
	ChecksumKey string
	PartnerCode string // Optional
	ReturnURL   string
	CancelURL   string
}

// NewService creates a new PayOS service with the official SDK
func NewService(config *Config) (*Service, error) {
	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("PAYOS_CLIENT_ID is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("PAYOS_API_KEY is required")
	}
	if config.ChecksumKey == "" {
		return nil, fmt.Errorf("PAYOS_CHECKSUM_KEY is required")
	}

	// Initialize PayOS with keys
	var err error
	if config.PartnerCode != "" {
		err = payossdk.Key(config.ClientID, config.APIKey, config.ChecksumKey, config.PartnerCode)
	} else {
		err = payossdk.Key(config.ClientID, config.APIKey, config.ChecksumKey)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize PayOS: %w", err)
	}

	return &Service{
		initialized: true,
		config:      config,
	}, nil
}

// CreatePaymentLink creates a new payment link using the official SDK
func (s *Service) CreatePaymentLink(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if !s.initialized {
		return nil, fmt.Errorf("PayOS service not initialized")
	}

	// Convert our items to PayOS format
	var items []payossdk.Item
	for _, item := range req.Items {
		items = append(items, payossdk.Item{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	// Create PayOS request
	paymentRequest := payossdk.CheckoutRequestType{
		OrderCode:   req.OrderCode,
		Amount:      req.Amount,
		Description: req.Description,
		Items:       items,
		ReturnUrl:   req.ReturnURL,
		CancelUrl:   req.CancelURL,
	}

	// Create payment link
	response, err := payossdk.CreatePaymentLink(paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment link: %w", err)
	}

	// Convert SDK response to our response format
	return &CreatePaymentResponse{
		Code:    "00", // Success code
		Desc:    "Thành công",
		Success: true,
		Data: PaymentData{
			Bin:           response.Bin,
			AccountNumber: response.AccountNumber,
			AccountName:   response.AccountName,
			Amount:        response.Amount,
			Description:   response.Description,
			OrderCode:     response.OrderCode,
			Currency:      response.Currency,
			PaymentLinkId: response.PaymentLinkId,
			Status:        response.Status,
			CheckoutUrl:   response.CheckoutUrl,
			QrCode:        response.QRCode,
		},
	}, nil
}

// GetPaymentLinkInformation retrieves payment information
func (s *Service) GetPaymentLinkInformation(ctx context.Context, orderCode int64) (*PaymentInfoResponse, error) {
	if !s.initialized {
		return nil, fmt.Errorf("PayOS service not initialized")
	}

	orderCodeStr := strconv.FormatInt(orderCode, 10)
	response, err := payossdk.GetPaymentLinkInformation(orderCodeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment information: %w", err)
	}

	// Convert SDK response to our response format
	return &PaymentInfoResponse{
		Code:    "00", // Success code
		Desc:    "Thành công",
		Success: true,
		Data: PaymentInfoData{
			OrderCode:       response.OrderCode,
			Amount:          response.Amount,
			AmountPaid:      response.AmountPaid,
			AmountRemaining: response.AmountRemaining,
			Status:          response.Status,
			CreatedAt:       response.CreateAt,      // SDK uses CreateAt not CreatedAt
			Transactions:    []PaymentTransaction{}, // Convert from SDK transactions if needed
		},
	}, nil
}

// CancelPaymentLink cancels a payment link
func (s *Service) CancelPaymentLink(ctx context.Context, orderCode int64, cancelReason string) error {
	if !s.initialized {
		return fmt.Errorf("PayOS service not initialized")
	}

	orderCodeStr := strconv.FormatInt(orderCode, 10)
	_, err := payossdk.CancelPaymentLink(orderCodeStr, &cancelReason)
	if err != nil {
		return fmt.Errorf("failed to cancel payment link: %w", err)
	}

	return nil
}

// VerifyPaymentWebhookData verifies webhook data signature
func (s *Service) VerifyPaymentWebhookData(webhookData payossdk.WebhookType) (*payossdk.WebhookDataType, error) {
	if !s.initialized {
		return nil, fmt.Errorf("PayOS service not initialized")
	}

	verifiedData, err := payossdk.VerifyPaymentWebhookData(webhookData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify webhook data: %w", err)
	}

	return verifiedData, nil
}

// GetPaymentStatus maps PayOS status to our internal status
func GetPaymentStatus(payosStatus string) aggregate.PaymentStatus {
	switch payosStatus {
	case "PAID":
		return aggregate.PaymentStatusPaid
	case "CANCELLED":
		return aggregate.PaymentStatusCancelled
	case "EXPIRED":
		return aggregate.PaymentStatusExpired
	case "PENDING":
		return aggregate.PaymentStatusPending
	default:
		return aggregate.PaymentStatusPending
	}
}

// CreateWebhookDataFromMap creates WebhookType from a map (useful for HTTP handlers)
func CreateWebhookDataFromMap(data map[string]interface{}) (*payossdk.WebhookType, error) {
	webhookType := &payossdk.WebhookType{}

	if code, ok := data["code"].(string); ok {
		webhookType.Code = code
	}

	if desc, ok := data["desc"].(string); ok {
		webhookType.Desc = desc
	}

	if success, ok := data["success"].(bool); ok {
		webhookType.Success = success
	}

	if signature, ok := data["signature"].(string); ok {
		webhookType.Signature = signature
	}

	// Handle the nested data object
	if dataObj, ok := data["data"].(map[string]interface{}); ok {
		webhookData := &payossdk.WebhookDataType{}

		if orderCode, ok := dataObj["orderCode"]; ok {
			switch v := orderCode.(type) {
			case float64:
				webhookData.OrderCode = int64(v)
			case int64:
				webhookData.OrderCode = v
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					webhookData.OrderCode = parsed
				}
			}
		}

		if amount, ok := dataObj["amount"]; ok {
			switch v := amount.(type) {
			case float64:
				webhookData.Amount = int(v)
			case int:
				webhookData.Amount = v
			}
		}

		if description, ok := dataObj["description"].(string); ok {
			webhookData.Description = description
		}

		if accountNumber, ok := dataObj["accountNumber"].(string); ok {
			webhookData.AccountNumber = accountNumber
		}

		if reference, ok := dataObj["reference"].(string); ok {
			webhookData.Reference = reference
		}

		if transactionDateTime, ok := dataObj["transactionDateTime"].(string); ok {
			webhookData.TransactionDateTime = transactionDateTime
		}

		if currency, ok := dataObj["currency"].(string); ok {
			webhookData.Currency = currency
		}

		if paymentLinkId, ok := dataObj["paymentLinkId"].(string); ok {
			webhookData.PaymentLinkId = paymentLinkId
		}

		if code, ok := dataObj["code"].(string); ok {
			webhookData.Code = code
		}

		if desc, ok := dataObj["desc"].(string); ok {
			webhookData.Desc = desc
		}

		webhookType.Data = webhookData
	}

	return webhookType, nil
}
