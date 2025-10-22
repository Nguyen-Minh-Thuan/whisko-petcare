package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/response"
)

// ScheduleController handles HTTP requests for schedule operations
type ScheduleController struct {
	service *services.ScheduleService
}

// NewScheduleController creates a new schedule controller
func NewScheduleController(service *services.ScheduleService) *ScheduleController {
	return &ScheduleController{
		service: service,
	}
}

// CreateSchedule handles POST /schedules
func (c *ScheduleController) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var cmd command.CreateSchedule
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	if err := c.service.CreateSchedule(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendCreated(w, r, nil)
}

// GetSchedule handles GET /schedules/{id}
func (c *ScheduleController) GetSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		response.SendBadRequest(w, r, "Schedule ID is required")
		return
	}

	schedule, err := c.service.GetSchedule(r.Context(), scheduleID)
	if err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, schedule)
}

// ListSchedules handles GET /schedules
func (c *ScheduleController) ListSchedules(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	schedules, err := c.service.ListSchedules(r.Context(), offset, limit)
	if err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, schedules)
}

// ListUserSchedules handles GET /users/{userID}/schedules
func (c *ScheduleController) ListUserSchedules(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	userID := r.PathValue("userID")
	if userID == "" {
		response.SendBadRequest(w, r, "User ID is required")
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	schedules, err := c.service.ListUserSchedules(r.Context(), userID, offset, limit)
	if err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, schedules)
}

// ListShopSchedules handles GET /vendors/{shopID}/schedules
func (c *ScheduleController) ListShopSchedules(w http.ResponseWriter, r *http.Request) {
	// Extract shop ID from path
	shopID := r.PathValue("shopID")
	if shopID == "" {
		response.SendBadRequest(w, r, "Shop ID is required")
		return
	}

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	schedules, err := c.service.ListShopSchedules(r.Context(), shopID, offset, limit)
	if err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, schedules)
}

// ChangeScheduleStatus handles PUT /schedules/{id}/status
func (c *ScheduleController) ChangeScheduleStatus(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		response.SendBadRequest(w, r, "Schedule ID is required")
		return
	}

	var cmd command.ChangeScheduleStatus
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Set schedule ID from path
	cmd.ScheduleID = scheduleID

	if err := c.service.ChangeScheduleStatus(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, nil)
}

// CompleteSchedule handles POST /schedules/{id}/complete
func (c *ScheduleController) CompleteSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		response.SendBadRequest(w, r, "Schedule ID is required")
		return
	}

	cmd := &command.CompleteSchedule{
		ScheduleID: scheduleID,
	}

	if err := c.service.CompleteSchedule(r.Context(), cmd); err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, nil)
}

// CancelSchedule handles POST /schedules/{id}/cancel
func (c *ScheduleController) CancelSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		response.SendBadRequest(w, r, "Schedule ID is required")
		return
	}

	var cmd command.CancelSchedule
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Set schedule ID from path
	cmd.ScheduleID = scheduleID

	if err := c.service.CancelSchedule(r.Context(), &cmd); err != nil {
		response.SendInternalError(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, nil)
}
