package services

import (
	"context"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
)

// ScheduleService handles schedule operations
type ScheduleService struct {
	createScheduleHandler        *command.CreateScheduleWithUoWHandler
	changeScheduleStatusHandler  *command.ChangeScheduleStatusWithUoWHandler
	completeScheduleHandler      *command.CompleteScheduleWithUoWHandler
	cancelScheduleHandler        *command.CancelScheduleWithUoWHandler
	getScheduleHandler           *query.GetScheduleHandler
	listUserSchedulesHandler     *query.ListUserSchedulesHandler
	listShopSchedulesHandler     *query.ListShopSchedulesHandler
	listSchedulesHandler         *query.ListSchedulesHandler
}

// NewScheduleService creates a new schedule service
func NewScheduleService(
	createScheduleHandler *command.CreateScheduleWithUoWHandler,
	changeScheduleStatusHandler *command.ChangeScheduleStatusWithUoWHandler,
	completeScheduleHandler *command.CompleteScheduleWithUoWHandler,
	cancelScheduleHandler *command.CancelScheduleWithUoWHandler,
	getScheduleHandler *query.GetScheduleHandler,
	listUserSchedulesHandler *query.ListUserSchedulesHandler,
	listShopSchedulesHandler *query.ListShopSchedulesHandler,
	listSchedulesHandler *query.ListSchedulesHandler,
) *ScheduleService {
	return &ScheduleService{
		createScheduleHandler:       createScheduleHandler,
		changeScheduleStatusHandler: changeScheduleStatusHandler,
		completeScheduleHandler:     completeScheduleHandler,
		cancelScheduleHandler:       cancelScheduleHandler,
		getScheduleHandler:          getScheduleHandler,
		listUserSchedulesHandler:    listUserSchedulesHandler,
		listShopSchedulesHandler:    listShopSchedulesHandler,
		listSchedulesHandler:        listSchedulesHandler,
	}
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, cmd *command.CreateSchedule) error {
	return s.createScheduleHandler.Handle(ctx, cmd)
}

// ChangeScheduleStatus changes the status of a schedule
func (s *ScheduleService) ChangeScheduleStatus(ctx context.Context, cmd *command.ChangeScheduleStatus) error {
	return s.changeScheduleStatusHandler.Handle(ctx, cmd)
}

// CompleteSchedule marks a schedule as completed
func (s *ScheduleService) CompleteSchedule(ctx context.Context, cmd *command.CompleteSchedule) error {
	return s.completeScheduleHandler.Handle(ctx, cmd)
}

// CancelSchedule cancels a schedule
func (s *ScheduleService) CancelSchedule(ctx context.Context, cmd *command.CancelSchedule) error {
	return s.cancelScheduleHandler.Handle(ctx, cmd)
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleService) GetSchedule(ctx context.Context, scheduleID string) (interface{}, error) {
	return s.getScheduleHandler.Handle(ctx, scheduleID)
}

// ListUserSchedules retrieves all schedules for a user with pagination
func (s *ScheduleService) ListUserSchedules(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	return s.listUserSchedulesHandler.Handle(ctx, userID, offset, limit)
}

// ListShopSchedules retrieves all schedules for a shop with pagination
func (s *ScheduleService) ListShopSchedules(ctx context.Context, shopID string, offset, limit int) ([]interface{}, error) {
	return s.listShopSchedulesHandler.Handle(ctx, shopID, offset, limit)
}

// ListSchedules retrieves all schedules with pagination
func (s *ScheduleService) ListSchedules(ctx context.Context, offset, limit int) ([]interface{}, error) {
	return s.listSchedulesHandler.Handle(ctx, offset, limit)
}
