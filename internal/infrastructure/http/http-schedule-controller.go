package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
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
	var req command.CreateSchedule

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	if err := c.service.CreateSchedule(r.Context(), &req); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Schedule created successfully",
	}
	response.SendCreated(w, r, responseData)
}

// GetSchedule handles GET /schedules/{id}
func (c *ScheduleController) GetSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Schedule ID is required"))
		return
	}

	schedule, err := c.service.GetSchedule(r.Context(), scheduleID)
	if err != nil {
		middleware.HandleError(w, r, err)
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
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, schedules)
}

// ListUserSchedules handles GET /users/{userID}/schedules
func (c *ScheduleController) ListUserSchedules(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("User ID is required"))
		return
	}
	userID := parts[0]

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Debug log
	println("📥 ListUserSchedules - Received request for userID:", userID, "offset:", offset, "limit:", limit)

	schedules, err := c.service.ListUserSchedules(r.Context(), userID, offset, limit)
	if err != nil {
		println("❌ ListUserSchedules - Error:", err.Error())
		middleware.HandleError(w, r, err)
		return
	}

	println("📤 ListUserSchedules - Returning schedules, count:", len(schedules))
	response.SendSuccess(w, r, schedules)
}

// ListShopSchedules handles GET /vendors/{shopID}/schedules
func (c *ScheduleController) ListShopSchedules(w http.ResponseWriter, r *http.Request) {
	// Extract shop ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/vendors/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Shop ID is required"))
		return
	}
	shopID := parts[0]

	// Parse query parameters
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	schedules, err := c.service.ListShopSchedules(r.Context(), shopID, offset, limit)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, schedules)
}

// ChangeScheduleStatus handles PUT /schedules/{id}/status
func (c *ScheduleController) ChangeScheduleStatus(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Schedule ID is required"))
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.ChangeScheduleStatus{
		ScheduleID: scheduleID,
		Status:     req.Status,
	}

	if err := c.service.ChangeScheduleStatus(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Schedule status changed successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// CompleteSchedule handles POST /schedules/{id}/complete
func (c *ScheduleController) CompleteSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Schedule ID is required"))
		return
	}

	cmd := &command.CompleteSchedule{
		ScheduleID: scheduleID,
	}

	if err := c.service.CompleteSchedule(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Schedule completed successfully",
	}
	response.SendSuccess(w, r, responseData)
}

// CancelSchedule handles POST /schedules/{id}/cancel
func (c *ScheduleController) CancelSchedule(w http.ResponseWriter, r *http.Request) {
	// Extract schedule ID from path
	scheduleID := r.PathValue("id")
	if scheduleID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("Schedule ID is required"))
		return
	}

	var req struct {
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CancelSchedule{
		ScheduleID: scheduleID,
		Reason:     req.Reason,
	}

	if err := c.service.CancelSchedule(r.Context(), &cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"message": "Schedule cancelled successfully",
	}
	response.SendSuccess(w, r, responseData)
}
