package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/infrastructure/payos"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/response"
)

// PaymentCommandHandlers defines interfaces for payment command handlers
type CreatePaymentHandlerInterface interface {
	Handle(ctx context.Context, cmd *command.CreatePaymentCommand) (*command.CreatePaymentResponse, error)
}

type CancelPaymentHandlerInterface interface {
	Handle(ctx context.Context, cmd *command.CancelPaymentCommand) error
}

type ConfirmPaymentHandlerInterface interface {
	Handle(ctx context.Context, cmd *command.ConfirmPaymentCommand) error
}

// HTTPPaymentController handles HTTP requests for payment operations
type HTTPPaymentController struct {
	createPaymentHandler         CreatePaymentHandlerInterface
	cancelPaymentHandler         CancelPaymentHandlerInterface
	confirmPaymentHandler        ConfirmPaymentHandlerInterface
	getPaymentHandler            *query.GetPaymentHandler
	getPaymentByOrderCodeHandler *query.GetPaymentByOrderCodeHandler
	listUserPaymentsHandler      *query.ListUserPaymentsHandler
	payOSService                 *payos.Service
}

// NewHTTPPaymentController creates a new HTTP payment controller
func NewHTTPPaymentController(
	createPaymentHandler CreatePaymentHandlerInterface,
	cancelPaymentHandler CancelPaymentHandlerInterface,
	confirmPaymentHandler ConfirmPaymentHandlerInterface,
	getPaymentHandler *query.GetPaymentHandler,
	getPaymentByOrderCodeHandler *query.GetPaymentByOrderCodeHandler,
	listUserPaymentsHandler *query.ListUserPaymentsHandler,
	payOSService *payos.Service,
) *HTTPPaymentController {
	return &HTTPPaymentController{
		createPaymentHandler:         createPaymentHandler,
		cancelPaymentHandler:         cancelPaymentHandler,
		confirmPaymentHandler:        confirmPaymentHandler,
		getPaymentHandler:            getPaymentHandler,
		getPaymentByOrderCodeHandler: getPaymentByOrderCodeHandler,
		listUserPaymentsHandler:      listUserPaymentsHandler,
		payOSService:                 payOSService,
	}
}

// CreatePayment handles POST /payments
func (c *HTTPPaymentController) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var cmd command.CreatePaymentCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	result, err := c.createPaymentHandler.Handle(r.Context(), &cmd)
	if err != nil {
		response.SendBadRequest(w, r, "Failed to create payment: "+err.Error())
		return
	}

	response.SendCreated(w, r, result)
}

// CheckAndUpdatePaymentStatus manually checks payment status with PayOS and updates it
// This is useful for local development when webhook is not accessible
func (c *HTTPPaymentController) CheckAndUpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	// Extract order code from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payments/check/")
	if path == "" {
		response.SendBadRequest(w, r, "Order code is required")
		return
	}

	orderCode, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		response.SendBadRequest(w, r, "Invalid order code")
		return
	}

	fmt.Printf("========================================\n")
	fmt.Printf("🔍 Manual payment status check requested for order: %d\n", orderCode)
	fmt.Printf("========================================\n")

	// Call confirm payment handler (it will check PayOS and update status)
	cmd := &command.ConfirmPaymentCommand{
		OrderCode: orderCode,
	}

	err = c.confirmPaymentHandler.Handle(r.Context(), cmd)
	if err != nil {
		fmt.Printf("❌ Failed to check/update payment: %v\n", err)
		response.SendInternalError(w, r, "Failed to check payment status: "+err.Error())
		return
	}

	fmt.Printf("✅ Payment status check completed\n")
	response.SendSuccess(w, r, map[string]string{
		"message": "Payment status checked and updated successfully",
	})
}

// GetPayment handles GET /payments/{id}
func (c *HTTPPaymentController) GetPayment(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payments/")
	if path == "" {
		response.SendBadRequest(w, r, "Payment ID is required")
		return
	}

	query := &query.GetPaymentQuery{PaymentID: path}
	payment, err := c.getPaymentHandler.Handle(r.Context(), query)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Payment not found")
			return
		}
		response.SendInternalError(w, r, "Failed to get payment")
		return
	}

	response.SendSuccess(w, r, c.paymentToResponse(payment))
}

// GetPaymentByOrderCode handles GET /payments/order/{orderCode}
func (c *HTTPPaymentController) GetPaymentByOrderCode(w http.ResponseWriter, r *http.Request) {
	// Extract order code from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payments/order/")
	if path == "" {
		response.SendBadRequest(w, r, "Order code is required")
		return
	}

	orderCode, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		response.SendBadRequest(w, r, "Invalid order code")
		return
	}

	query := &query.GetPaymentByOrderCodeQuery{OrderCode: orderCode}
	payment, err := c.getPaymentByOrderCodeHandler.Handle(r.Context(), query)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Payment not found")
			return
		}
		response.SendInternalError(w, r, "Failed to get payment")
		return
	}

	response.SendSuccess(w, r, c.paymentToResponse(payment))
}

// ListUserPayments handles GET /payments/user/{userID}
func (c *HTTPPaymentController) ListUserPayments(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payments/user/")
	if path == "" {
		response.SendBadRequest(w, r, "User ID is required")
		return
	}

	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 10 // default limit

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	query := &query.ListUserPaymentsQuery{
		UserID: path,
		Offset: offset,
		Limit:  limit,
	}

	payments, err := c.listUserPaymentsHandler.Handle(r.Context(), query)
	if err != nil {
		response.SendInternalError(w, r, "Failed to list payments")
		return
	}

	// Convert to response format
	paymentResponses := make([]map[string]interface{}, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = c.paymentToResponse(payment)
	}

	responseData := map[string]interface{}{
		"payments": paymentResponses,
		"offset":   offset,
		"limit":    limit,
		"count":    len(payments),
	}

	response.SendSuccess(w, r, responseData)
}

// CancelPayment handles PUT /payments/{id}/cancel
func (c *HTTPPaymentController) CancelPayment(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payments/")
	path = strings.TrimSuffix(path, "/cancel")
	if path == "" {
		response.SendBadRequest(w, r, "Payment ID is required")
		return
	}

	var cancelReq struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cancelReq); err != nil {
		// Allow empty body, reason is optional
		cancelReq.Reason = "Cancelled by user"
	}

	cmd := &command.CancelPaymentCommand{
		PaymentID: path,
		Reason:    cancelReq.Reason,
	}

	err := c.cancelPaymentHandler.Handle(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.SendNotFound(w, r, "Payment not found")
			return
		}
		response.SendBadRequest(w, r, "Failed to cancel payment: "+err.Error())
		return
	}

	response.SendSuccess(w, r, nil)
}

// WebhookHandler handles PayOS webhooks
func (c *HTTPPaymentController) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("========================================\n")
	fmt.Printf("🔔 WEBHOOK RECEIVED from PayOS\n")
	fmt.Printf("========================================\n")
	
	// Verify webhook signature
	signature := r.Header.Get("x-signature")
	if signature == "" {
		fmt.Printf("⚠️ Webhook received without signature (PayOS test request)\n")
		// Allow PayOS test requests during setup - respond with success
		response.SendSuccess(w, r, map[string]interface{}{
			"message": "Webhook endpoint is ready",
			"status": "ok",
		})
		return
	}
	fmt.Printf("✅ Signature present: %s...\n", signature[:20])

	// Read the body as a map first
	var webhookPayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&webhookPayload); err != nil {
		fmt.Printf("❌ Webhook rejected: Invalid payload - %v\n", err)
		response.SendBadRequest(w, r, "Invalid webhook payload")
		return
	}
	fmt.Printf("📦 Webhook payload received: %+v\n", webhookPayload)

	// Convert to PayOS webhook type
	webhookData, err := payos.CreateWebhookDataFromMap(webhookPayload)
	if err != nil {
		fmt.Printf("❌ Webhook rejected: Invalid data format - %v\n", err)
		response.SendBadRequest(w, r, "Invalid webhook data format")
		return
	}

	// Verify webhook data using PayOS SDK
	verifiedData, err := c.payOSService.VerifyPaymentWebhookData(*webhookData)
	if err != nil {
		fmt.Printf("❌ Webhook rejected: Verification failed - %v\n", err)
		response.SendBadRequest(w, r, "Webhook verification failed")
		return
	}
	fmt.Printf("✅ Webhook verified! Order Code: %d\n", verifiedData.OrderCode)

	// Process the webhook
	cmd := &command.ConfirmPaymentCommand{
		OrderCode: verifiedData.OrderCode,
	}

	fmt.Printf("🔄 Processing webhook for order: %d\n", verifiedData.OrderCode)
	err = c.confirmPaymentHandler.Handle(r.Context(), cmd)
	if err != nil {
		fmt.Printf("❌ Webhook processing failed: %v\n", err)
		fmt.Printf("========================================\n")
		response.SendInternalError(w, r, "Failed to process webhook")
		return
	}

	fmt.Printf("✅ Webhook processed successfully!\n")
	fmt.Printf("========================================\n")
	response.SendSuccess(w, r, nil)
}

// ReturnHandler handles PayOS return URL
func (c *HTTPPaymentController) ReturnHandler(w http.ResponseWriter, r *http.Request) {
	orderCode := r.URL.Query().Get("orderCode")
	if orderCode == "" {
		http.Error(w, "Missing order code", http.StatusBadRequest)
		return
	}

	orderCodeInt, err := strconv.ParseInt(orderCode, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order code", http.StatusBadRequest)
		return
	}

	// Get payment by order code to show status
	query := &query.GetPaymentByOrderCodeQuery{OrderCode: orderCodeInt}
	payment, err := c.getPaymentByOrderCodeHandler.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	// Return a simple HTML page showing payment status
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Payment Status</title>
		<meta charset="UTF-8">
		<style>
			body { font-family: Arial, sans-serif; margin: 50px; text-align: center; }
			.success { color: green; }
			.pending { color: orange; }
			.error { color: red; }
		</style>
	</head>
	<body>
		<h1>Payment Status</h1>
		<p><strong>Order Code:</strong> %d</p>
		<p><strong>Amount:</strong> %d VND</p>
		<p><strong>Status:</strong> <span class="%s">%s</span></p>
		<p><strong>Description:</strong> %s</p>
		<hr>
		<p><a href="/">Return to Homepage</a></p>
	</body>
	</html>`,
		payment.OrderCode,
		payment.Amount,
		strings.ToLower(payment.Status),
		payment.Status,
		payment.Description,
	)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// CancelHandler handles PayOS cancel URL
func (c *HTTPPaymentController) CancelHandler(w http.ResponseWriter, r *http.Request) {
	orderCode := r.URL.Query().Get("orderCode")
	if orderCode == "" {
		http.Error(w, "Missing order code", http.StatusBadRequest)
		return
	}

	orderCodeInt, err := strconv.ParseInt(orderCode, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order code", http.StatusBadRequest)
		return
	}

	// Get payment by order code
	query := &query.GetPaymentByOrderCodeQuery{OrderCode: orderCodeInt}
	payment, err := c.getPaymentByOrderCodeHandler.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, "Payment not found", http.StatusNotFound)
		return
	}

	// Cancel the payment if it's still pending
	if payment.Status == "PENDING" {
		cancelCmd := &command.CancelPaymentCommand{
			PaymentID: payment.ID,
			Reason:    "Cancelled by user via cancel URL",
		}
		c.cancelPaymentHandler.Handle(r.Context(), cancelCmd)
	}

	// Return a simple HTML page
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Payment Cancelled</title>
		<meta charset="UTF-8">
		<style>
			body { font-family: Arial, sans-serif; margin: 50px; text-align: center; }
			.cancelled { color: red; }
		</style>
	</head>
	<body>
		<h1>Payment Cancelled</h1>
		<p><strong>Order Code:</strong> %d</p>
		<p><strong>Amount:</strong> %d VND</p>
		<p><span class="cancelled">Your payment has been cancelled.</span></p>
		<hr>
		<p><a href="/">Return to Homepage</a></p>
	</body>
	</html>`,
		payment.OrderCode,
		payment.Amount,
	)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// paymentToResponse converts a PaymentReadModel to a response map
func (c *HTTPPaymentController) paymentToResponse(payment *projection.PaymentReadModel) map[string]interface{} {
	return map[string]interface{}{
		"id":                   payment.ID,
		"order_code":           payment.OrderCode,
		"user_id":              payment.UserID,
		"amount":               payment.Amount,
		"description":          payment.Description,
		"items":                payment.Items,
		"status":               payment.Status,
		"method":               payment.Method,
		"payos_transaction_id": payment.PayOSTransactionID,
		"checkout_url":         payment.CheckoutURL,
		"qr_code":              payment.QRCode,
		"expired_at":           payment.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
		"created_at":           payment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":           payment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
