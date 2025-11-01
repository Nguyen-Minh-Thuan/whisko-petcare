package command

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/internal/infrastructure/payos"
	"whisko-petcare/pkg/errors"
)

// CreatePaymentWithUoWHandler handles create payment commands with Unit of Work
type CreatePaymentWithUoWHandler struct {
	uowFactory   repository.UnitOfWorkFactory
	eventBus     bus.EventBus
	payOSService *payos.Service
}

// NewCreatePaymentWithUoWHandler creates a new create payment handler with UoW
func NewCreatePaymentWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
	payOSService *payos.Service,
) *CreatePaymentWithUoWHandler {
	return &CreatePaymentWithUoWHandler{
		uowFactory:   uowFactory,
		eventBus:     eventBus,
		payOSService: payOSService,
	}
}

// Handle processes the create payment command
func (h *CreatePaymentWithUoWHandler) Handle(ctx context.Context, cmd *CreatePaymentCommand) (*CreatePaymentResponse, error) {
	if cmd == nil {
		return nil, errors.NewValidationError("command cannot be nil")
	}

	// Validate request
	if cmd.UserID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}
	if cmd.Amount <= 0 {
		return nil, errors.NewValidationError("amount must be greater than 0")
	}
	if cmd.Description == "" {
		return nil, errors.NewValidationError("description is required")
	}
	if len(cmd.Items) == 0 {
		return nil, errors.NewValidationError("items are required")
	}
	if cmd.VendorID == "" {
		return nil, errors.NewValidationError("vendor_id is required")
	}
	if cmd.PetID == "" {
		return nil, errors.NewValidationError("pet_id is required")
	}
	if len(cmd.ServiceIDs) == 0 {
		return nil, errors.NewValidationError("service_ids are required")
	}
	if cmd.StartTime == "" {
		return nil, errors.NewValidationError("start_time is required")
	}
	if cmd.EndTime == "" {
		return nil, errors.NewValidationError("end_time is required")
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, cmd.StartTime)
	if err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid start_time format: %v", err))
	}
	endTime, err := time.Parse(time.RFC3339, cmd.EndTime)
	if err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid end_time format: %v", err))
	}

	// Validate that item total matches amount
	totalAmount := 0
	for _, item := range cmd.Items {
		totalAmount += item.Price * item.Quantity
	}
	if totalAmount != cmd.Amount {
		return nil, errors.NewValidationError(fmt.Sprintf("total item amount (%d) does not match payment amount (%d)", totalAmount, cmd.Amount))
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Create payment aggregate with schedule information
	payment, err := aggregate.NewPayment(cmd.UserID, cmd.Amount, cmd.Description, cmd.Items, cmd.VendorID, cmd.PetID, cmd.ServiceIDs, startTime, endTime)
	if err != nil {
		uow.Rollback(ctx)
		return nil, errors.NewValidationError(fmt.Sprintf("failed to create payment: %v", err))
	}

	// Convert items for PayOS API
	payOSItems := make([]payos.PaymentItem, len(cmd.Items))
	for i, item := range cmd.Items {
		payOSItems[i] = payos.PaymentItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	// Create payment request for PayOS
	// PayOS requires description to be max 25 characters
	description := cmd.Description
	if len(description) > 25 {
		description = description[:25]
	}
	
	payOSReq := &payos.CreatePaymentRequest{
		OrderCode:   payment.OrderCode(),
		Amount:      cmd.Amount,
		Description: description,
		Items:       payOSItems,
		ReturnURL:   h.payOSService.GetReturnURL(),
		CancelURL:   h.payOSService.GetCancelURL(),
	}

	// Create payment in PayOS using the service
	payOSResp, err := h.payOSService.CreatePaymentLink(ctx, payOSReq)
	if err != nil {
		uow.Rollback(ctx)
		return nil, errors.NewInternalError(fmt.Sprintf("failed to create PayOS payment: %v", err))
	}

	if !payOSResp.Success {
		uow.Rollback(ctx)
		return nil, errors.NewInternalError(fmt.Sprintf("PayOS payment creation failed: %s", payOSResp.Desc))
	}

	// Update payment with PayOS details
	err = payment.SetPayOSDetails(
		payOSResp.Data.PaymentLinkId,
		payOSResp.Data.CheckoutUrl,
		payOSResp.Data.QrCode,
	)
	if err != nil {
		uow.Rollback(ctx)
		return nil, errors.NewInternalError(fmt.Sprintf("failed to set PayOS details: %v", err))
	}

	// Save payment using repository from unit of work
	paymentRepo := uow.PaymentRepository()
	if err := paymentRepo.Save(ctx, payment); err != nil {
		uow.Rollback(ctx)
		return nil, errors.NewInternalError(fmt.Sprintf("failed to save payment: %v", err))
	}

	// Publish events asynchronously
	events := payment.GetUncommittedEvents()
	fmt.Printf("DEBUG: Publishing %d payment events\n", len(events))
	for i, evt := range events {
		fmt.Printf("DEBUG: Event %d - Type: %T\n", i, evt)
	}
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		// Log warning but don't fail the command (eventual consistency)
		fmt.Printf("ERROR: failed to publish payment events: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Successfully published %d events\n", len(events))
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return &CreatePaymentResponse{
		PaymentID:   payment.ID(),
		OrderCode:   payment.OrderCode(),
		CheckoutURL: payOSResp.Data.CheckoutUrl,
		QRCode:      payOSResp.Data.QrCode,
		Amount:      payment.Amount(),
		Status:      string(payment.Status()),
		ExpiredAt:   payment.ExpiredAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CancelPaymentWithUoWHandler handles cancel payment commands with Unit of Work
type CancelPaymentWithUoWHandler struct {
	uowFactory   repository.UnitOfWorkFactory
	eventBus     bus.EventBus
	payOSService *payos.Service
}

// NewCancelPaymentWithUoWHandler creates a new cancel payment handler with UoW
func NewCancelPaymentWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
	payOSService *payos.Service,
) *CancelPaymentWithUoWHandler {
	return &CancelPaymentWithUoWHandler{
		uowFactory:   uowFactory,
		eventBus:     eventBus,
		payOSService: payOSService,
	}
}

// Handle processes the cancel payment command
func (h *CancelPaymentWithUoWHandler) Handle(ctx context.Context, cmd *CancelPaymentCommand) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	if cmd.PaymentID == "" {
		return errors.NewValidationError("payment_id is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get payment from repository
	paymentRepo := uow.PaymentRepository()
	payment, err := paymentRepo.GetByID(ctx, cmd.PaymentID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("payment not found: %v", err))
	}

	reason := cmd.Reason
	if reason == "" {
		reason = "Cancelled by user"
	}

	// Cancel in PayOS if still pending
	if payment.Status() == aggregate.PaymentStatusPending {
		err = h.payOSService.CancelPaymentLink(ctx, payment.OrderCode(), reason)
		if err != nil {
			uow.Rollback(ctx)
			return errors.NewInternalError(fmt.Sprintf("failed to cancel PayOS payment: %v", err))
		}
	}

	// Mark payment as cancelled
	err = payment.MarkAsCancelled()
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to mark payment as cancelled: %v", err))
	}

	// Save updated payment
	if err := paymentRepo.Save(ctx, payment); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save payment: %v", err))
	}

	// Publish events asynchronously
	events := payment.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish payment events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// ConfirmPaymentWithUoWHandler handles confirm payment commands with Unit of Work
type ConfirmPaymentWithUoWHandler struct {
	uowFactory              repository.UnitOfWorkFactory
	eventBus                bus.EventBus
	payOSService            *payos.Service
	createScheduleHandler   *CreateScheduleWithUoWHandler
}

// NewConfirmPaymentWithUoWHandler creates a new confirm payment handler with UoW
func NewConfirmPaymentWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
	payOSService *payos.Service,
	createScheduleHandler *CreateScheduleWithUoWHandler,
) *ConfirmPaymentWithUoWHandler {
	return &ConfirmPaymentWithUoWHandler{
		uowFactory:              uowFactory,
		eventBus:                eventBus,
		payOSService:            payOSService,
		createScheduleHandler:   createScheduleHandler,
	}
}

// Handle processes the confirm payment command
func (h *ConfirmPaymentWithUoWHandler) Handle(ctx context.Context, cmd *ConfirmPaymentCommand) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	if cmd.OrderCode == 0 {
		return errors.NewValidationError("order_code is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get payment by order code from repository
	paymentRepo := uow.PaymentRepository()
	payment, err := paymentRepo.GetByOrderCode(ctx, cmd.OrderCode)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("payment not found: %v", err))
	}

	// Verify payment status with PayOS
	payOSInfo, err := h.payOSService.GetPaymentLinkInformation(ctx, cmd.OrderCode)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to get PayOS payment info: %v", err))
	}

	if !payOSInfo.Success {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("PayOS payment info request failed: %s", payOSInfo.Desc))
	}

	// Update payment status based on PayOS response
	var paymentWasPaid bool
	switch payOSInfo.Data.Status {
	case "PAID":
		err = payment.MarkAsPaid()
		paymentWasPaid = true
	case "CANCELLED":
		err = payment.MarkAsCancelled()
	case "EXPIRED":
		err = payment.MarkAsExpired()
	default:
		// Payment is still pending or in unknown state
		uow.Rollback(ctx)
		return nil
	}

	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update payment status: %v", err))
	}

	// Save updated payment
	if err := paymentRepo.Save(ctx, payment); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save payment: %v", err))
	}

	// Publish events asynchronously
	events := payment.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish payment events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	// AUTO-CREATE SCHEDULE: If payment was successful, automatically create a schedule
	if paymentWasPaid && h.createScheduleHandler != nil {
		fmt.Printf("Payment successful - auto-creating schedule for payment ID: %s\n", payment.ID())
		
		scheduleCmd := &CreateSchedule{
			UserID:    payment.UserID(),
			VendorID:  payment.VendorID(),
			PetID:     payment.PetID(),
			ServiceIDs: payment.ServiceIDs(),
			StartTime: payment.StartTime().Format(time.RFC3339),
			EndTime:   payment.EndTime().Format(time.RFC3339),
		}
		
		if err := h.createScheduleHandler.Handle(ctx, scheduleCmd); err != nil {
			// Log error but don't fail the payment confirmation
			fmt.Printf("Warning: failed to auto-create schedule after payment: %v\n", err)
		} else {
			fmt.Printf("Successfully auto-created schedule for payment ID: %s\n", payment.ID())
		}
	}

	return nil
}
