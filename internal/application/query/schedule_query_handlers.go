package query

import (
	"context"
	"fmt"

	"whisko-petcare/pkg/errors"
)

// ScheduleProjection interface for schedule read model
type ScheduleProjection interface {
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]interface{}, error)
	GetByShopID(ctx context.Context, shopID string, offset, limit int) ([]interface{}, error)
	ListAll(ctx context.Context, offset, limit int) ([]interface{}, error)
}

// GetScheduleHandler handles get schedule by ID queries
type GetScheduleHandler struct {
	projection ScheduleProjection
}

// NewGetScheduleHandler creates a new get schedule handler
func NewGetScheduleHandler(projection ScheduleProjection) *GetScheduleHandler {
	return &GetScheduleHandler{
		projection: projection,
	}
}

// Handle processes the get schedule query
func (h *GetScheduleHandler) Handle(ctx context.Context, scheduleID string) (interface{}, error) {
	if scheduleID == "" {
		return nil, errors.NewValidationError("schedule_id is required")
	}

	schedule, err := h.projection.GetByID(ctx, scheduleID)
	if err != nil {
		return nil, errors.NewNotFoundError("schedule")
	}

	return schedule, nil
}

// ListUserSchedulesHandler handles list user schedules queries
type ListUserSchedulesHandler struct {
	projection ScheduleProjection
}

// NewListUserSchedulesHandler creates a new list user schedules handler
func NewListUserSchedulesHandler(projection ScheduleProjection) *ListUserSchedulesHandler {
	return &ListUserSchedulesHandler{
		projection: projection,
	}
}

// Handle processes the list user schedules query
func (h *ListUserSchedulesHandler) Handle(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user_id is required")
	}

	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	schedules, err := h.projection.GetByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list user schedules: %v", err))
	}

	return schedules, nil
}

// ListShopSchedulesHandler handles list shop schedules queries
type ListShopSchedulesHandler struct {
	projection ScheduleProjection
}

// NewListShopSchedulesHandler creates a new list shop schedules handler
func NewListShopSchedulesHandler(projection ScheduleProjection) *ListShopSchedulesHandler {
	return &ListShopSchedulesHandler{
		projection: projection,
	}
}

// Handle processes the list shop schedules query
func (h *ListShopSchedulesHandler) Handle(ctx context.Context, shopID string, offset, limit int) ([]interface{}, error) {
	if shopID == "" {
		return nil, errors.NewValidationError("shop_id is required")
	}

	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	schedules, err := h.projection.GetByShopID(ctx, shopID, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list shop schedules: %v", err))
	}

	return schedules, nil
}

// ListSchedulesHandler handles list all schedules queries
type ListSchedulesHandler struct {
	projection ScheduleProjection
}

// NewListSchedulesHandler creates a new list schedules handler
func NewListSchedulesHandler(projection ScheduleProjection) *ListSchedulesHandler {
	return &ListSchedulesHandler{
		projection: projection,
	}
}

// Handle processes the list schedules query
func (h *ListSchedulesHandler) Handle(ctx context.Context, offset, limit int) ([]interface{}, error) {
	// Set default limit if not provided
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	schedules, err := h.projection.ListAll(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to list schedules: %v", err))
	}

	return schedules, nil
}
