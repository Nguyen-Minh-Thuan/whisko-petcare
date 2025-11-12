package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/infrastructure/mongo"
	"whisko-petcare/internal/infrastructure/payos"
	"whisko-petcare/pkg/response"
)

// HTTPPayoutController handles HTTP requests for payout operations
type HTTPPayoutController struct {
	uowFactory    *mongo.MongoUnitOfWorkFactory
	payoutService *payos.PayoutService
}

// NewHTTPPayoutController creates a new HTTP payout controller
func NewHTTPPayoutController(
	uowFactory *mongo.MongoUnitOfWorkFactory,
	payoutService *payos.PayoutService,
) *HTTPPayoutController {
	return &HTTPPayoutController{
		uowFactory:    uowFactory,
		payoutService: payoutService,
	}
}

// ProcessPayout handles POST /payouts/{id}/process
// This endpoint initiates a real bank transfer for the payout
func (c *HTTPPayoutController) ProcessPayout(w http.ResponseWriter, r *http.Request) {
	// Extract payout ID from URL path
	// The route is registered as /payouts/ not /api/payouts/
	fmt.Printf("🔍 ProcessPayout - Original URL Path: %s\n", r.URL.Path)
	path := strings.TrimPrefix(r.URL.Path, "/payouts/")
	fmt.Printf("🔍 ProcessPayout - After TrimPrefix: %s\n", path)
	path = strings.TrimSuffix(path, "/process")
	payoutID := path
	fmt.Printf("🔍 ProcessPayout - Final Payout ID: %s\n", payoutID)

	if payoutID == "" {
		response.SendBadRequest(w, r, "Payout ID is required")
		return
	}

	// Start Unit of Work
	uow := c.uowFactory.CreateUnitOfWork()
	defer uow.Rollback(r.Context())

	// Get payout aggregate
	payoutRepo := uow.PayoutRepository()
	payout, err := payoutRepo.GetByID(r.Context(), payoutID)
	if err != nil {
		response.SendNotFound(w, r, "Payout not found")
		return
	}

	// Check if payout is in pending or failed status (allow retry of failed payouts)
	if payout.Status() != aggregate.PayoutStatusPending && payout.Status() != aggregate.PayoutStatusFailed {
		response.SendBadRequest(w, r, "Payout can only be processed when status is PENDING or FAILED. Current status: "+string(payout.Status()))
		return
	}

	// Get vendor information
	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(r.Context(), payout.VendorID())
	if err != nil {
		response.SendInternalError(w, r, "Failed to retrieve vendor: "+err.Error())
		return
	}

	// Check if vendor has bank account
	if !vendor.HasBankAccount() {
		response.SendBadRequest(w, r, "Vendor does not have a bank account configured")
		return
	}

	bankAccount := vendor.GetBankAccount()

	// Now process the actual bank transfer FIRST to get transfer ID
	transferInfo, transferErr := c.payoutService.ProcessPayout(
		r.Context(),
		payout.ID(),
		bankAccount.BankName,
		bankAccount.AccountNumber,
		bankAccount.AccountName,
		payout.Amount(),
		payout.Notes(),
	)

	// If transfer initiation failed immediately
	if transferErr != nil {
		if err := payout.MarkAsFailed(transferErr.Error()); err != nil {
			response.SendInternalError(w, r, "Failed to mark payout as failed: "+err.Error())
			return
		}

		if err := payoutRepo.Save(r.Context(), payout); err != nil {
			response.SendInternalError(w, r, "Failed to save failed payout: "+err.Error())
			return
		}

		if err := uow.Commit(r.Context()); err != nil {
			response.SendInternalError(w, r, "Failed to commit failed status: "+err.Error())
			return
		}

		response.SendBadRequest(w, r, "Payout transfer failed: "+transferErr.Error())
		return
	}

	// Mark payout as processing with transfer ID
	if err := payout.MarkAsProcessing(transferInfo.TransferID); err != nil {
		response.SendBadRequest(w, r, "Failed to mark payout as processing: "+err.Error())
		return
	}

	// Update payout based on transfer result
	if transferInfo.Status == "SUCCEEDED" {
		if err := payout.MarkAsCompleted(); err != nil {
			response.SendInternalError(w, r, "Failed to mark payout as completed: "+err.Error())
			return
		}
	} else if transferInfo.Status == "FAILED" {
		errorMsg := "Transfer failed"
		if transferInfo.ErrorMessage != "" {
			errorMsg = transferInfo.ErrorMessage
		}
		if err := payout.MarkAsFailed(errorMsg); err != nil {
			response.SendInternalError(w, r, "Failed to mark payout as failed: "+err.Error())
			return
		}
	}
	// If status is PROCESSING, we keep it as PROCESSING and wait for webhook

	// Save final status
	if err := payoutRepo.Save(r.Context(), payout); err != nil {
		response.SendInternalError(w, r, "Failed to save payout: "+err.Error())
		return
	}

	// Commit the transaction
	if err := uow.Commit(r.Context()); err != nil {
		response.SendInternalError(w, r, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return success response
	response.SendSuccess(w, r, map[string]interface{}{
		"payoutId":   payout.ID(),
		"transferId": transferInfo.TransferID,
		"status":     payout.Status(),
		"amount":     payout.Amount(),
		"message":    "Payout processed successfully",
	})
}

// GetPayoutByID handles GET /api/payouts/{id}
func (c *HTTPPayoutController) GetPayoutByID(w http.ResponseWriter, r *http.Request) {
	// Extract payout ID from URL path
	// The route is registered as /payouts/ not /api/payouts/
	fmt.Printf("🔍 GetPayoutByID - Original URL Path: %s\n", r.URL.Path)
	path := strings.TrimPrefix(r.URL.Path, "/payouts/")
	payoutID := path
	fmt.Printf("🔍 GetPayoutByID - Extracted Payout ID: %s\n", payoutID)

	if payoutID == "" {
		response.SendBadRequest(w, r, "Payout ID is required")
		return
	}

	// Start Unit of Work for read
	uow := c.uowFactory.CreateUnitOfWork()
	defer uow.Rollback(r.Context())

	// Get payout
	payoutRepo := uow.PayoutRepository()
	payout, err := payoutRepo.GetByID(r.Context(), payoutID)
	if err != nil {
		response.SendNotFound(w, r, "Payout not found")
		return
	}

	// Get vendor information for response
	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(r.Context(), payout.VendorID())
	if err != nil {
		response.SendInternalError(w, r, "Failed to retrieve vendor: "+err.Error())
		return
	}

	// Return payout details
	response.SendSuccess(w, r, map[string]interface{}{
		"id":              payout.ID(),
		"vendorId":        payout.VendorID(),
		"vendorName":      vendor.Name(),
		"scheduleId":      payout.ScheduleID(),
		"paymentId":       payout.PaymentID(),
		"amount":          payout.Amount(),
		"status":          payout.Status(),
		"notes":           payout.Notes(),
		"payosTransferId": payout.PayosTransferID(),
		"failureReason":   payout.FailureReason(),
		"bankAccount":     payout.BankAccount(),
		"requestedAt":     payout.RequestedAt(),
		"processedAt":     payout.ProcessedAt(),
		"completedAt":     payout.CompletedAt(),
		"createdAt":       payout.CreatedAt(),
		"updatedAt":       payout.UpdatedAt(),
	})
}

// ListPayoutsByVendor handles GET /payouts/vendor/{vendorId}
func (c *HTTPPayoutController) ListPayoutsByVendor(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payouts/vendor/")
	vendorID := path

	if vendorID == "" {
		response.SendBadRequest(w, r, "Vendor ID is required")
		return
	}

	// Start Unit of Work for read
	uow := c.uowFactory.CreateUnitOfWork()
	defer uow.Rollback(r.Context())

	// Get payouts (with default pagination: limit=100, skip=0)
	payoutRepo := uow.PayoutRepository()
	payouts, err := payoutRepo.GetByVendorID(r.Context(), vendorID, 100, 0)
	if err != nil {
		response.SendInternalError(w, r, "Failed to get payouts: "+err.Error())
		return
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, payout := range payouts {
		results = append(results, map[string]interface{}{
			"id":         payout.ID(),
			"vendorId":   payout.VendorID(),
			"scheduleId": payout.ScheduleID(),
			"paymentId":  payout.PaymentID(),
			"amount":     payout.Amount(),
			"status":     payout.Status(),
			"notes":      payout.Notes(),
			"createdAt":  payout.CreatedAt(),
		})
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"payouts": results,
		"total":   len(results),
	})
}

// ListPayoutsByStatus handles GET /payouts/status/{status}
func (c *HTTPPayoutController) ListPayoutsByStatus(w http.ResponseWriter, r *http.Request) {
	// Extract status from URL path
	path := strings.TrimPrefix(r.URL.Path, "/payouts/status/")
	status := strings.ToUpper(path)

	if status == "" {
		response.SendBadRequest(w, r, "Status is required")
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"PENDING":    true,
		"PROCESSING": true,
		"COMPLETED":  true,
		"FAILED":     true,
	}
	if !validStatuses[status] {
		response.SendBadRequest(w, r, "Invalid status. Must be one of: PENDING, PROCESSING, COMPLETED, FAILED")
		return
	}

	// Start Unit of Work for read
	uow := c.uowFactory.CreateUnitOfWork()
	defer uow.Rollback(r.Context())

	// Get payouts (with default pagination: limit=100, skip=0)
	payoutRepo := uow.PayoutRepository()
	payouts, err := payoutRepo.GetByStatus(r.Context(), aggregate.PayoutStatus(status), 100, 0)
	if err != nil {
		response.SendInternalError(w, r, "Failed to get payouts: "+err.Error())
		return
	}

	// Convert to response format
	var results []map[string]interface{}
	for _, payout := range payouts {
		results = append(results, map[string]interface{}{
			"id":         payout.ID(),
			"vendorId":   payout.VendorID(),
			"scheduleId": payout.ScheduleID(),
			"paymentId":  payout.PaymentID(),
			"amount":     payout.Amount(),
			"status":     payout.Status(),
			"notes":      payout.Notes(),
			"createdAt":  payout.CreatedAt(),
		})
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"payouts": results,
		"total":   len(results),
		"status":  status,
	})
}

// WebhookHandler handles POST /api/payouts/webhook
// This receives transfer status updates from PayOS
func (c *HTTPPayoutController) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var webhookData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
		response.SendBadRequest(w, r, "Invalid webhook payload")
		return
	}

	// TODO: Verify webhook signature for security
	// signature := r.Header.Get("x-signature")

	// Extract payout reference ID from webhook
	referenceID, ok := webhookData["referenceId"].(string)
	if !ok {
		response.SendBadRequest(w, r, "Missing referenceId in webhook")
		return
	}

	// Extract status
	status, ok := webhookData["status"].(string)
	if !ok {
		response.SendBadRequest(w, r, "Missing status in webhook")
		return
	}

	// Start Unit of Work
	uow := c.uowFactory.CreateUnitOfWork()
	defer uow.Rollback(r.Context())

	// Get payout by ID (referenceID is the payout ID)
	payoutRepo := uow.PayoutRepository()
	payout, err := payoutRepo.GetByID(r.Context(), referenceID)
	if err != nil {
		response.SendNotFound(w, r, "Payout not found")
		return
	}

	// Update payout status based on webhook
	switch status {
	case "SUCCEEDED":
		if err := payout.MarkAsCompleted(); err != nil {
			response.SendInternalError(w, r, "Failed to mark payout as completed: "+err.Error())
			return
		}

	case "FAILED":
		errorMsg := "Transfer failed"
		if msg, ok := webhookData["errorMessage"].(string); ok && msg != "" {
			errorMsg = msg
		}
		if err := payout.MarkAsFailed(errorMsg); err != nil {
			response.SendInternalError(w, r, "Failed to mark payout as failed: "+err.Error())
			return
		}
	}

	// Save updated payout
	if err := payoutRepo.Save(r.Context(), payout); err != nil {
		response.SendInternalError(w, r, "Failed to save payout: "+err.Error())
		return
	}

	if err := uow.Commit(r.Context()); err != nil {
		response.SendInternalError(w, r, "Failed to commit transaction: "+err.Error())
		return
	}

	// Return 200 OK to acknowledge webhook
	response.SendSuccess(w, r, map[string]string{
		"status": "success",
	})
}
