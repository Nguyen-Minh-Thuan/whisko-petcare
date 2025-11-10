package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

type HTTPAdminDashboardController struct {
	dashboardHandler *query.AdminDashboardHandler
}

func NewHTTPAdminDashboardController(dashboardHandler *query.AdminDashboardHandler) *HTTPAdminDashboardController {
	return &HTTPAdminDashboardController{
		dashboardHandler: dashboardHandler,
	}
}

// GetDashboardStats handles GET /admin/dashboard
// Query parameters: from_date, to_date (format: 2006-01-02 or RFC3339)
func (c *HTTPAdminDashboardController) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate time.Time
	var err error

	// Parse from_date
	if fromDateStr != "" {
		fromDate, err = parseDate(fromDateStr)
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError(fmt.Sprintf("invalid from_date format: %v", err)))
			return
		}
	} else {
		// Default to 30 days ago
		fromDate = time.Now().AddDate(0, 0, -30)
	}

	// Parse to_date
	if toDateStr != "" {
		toDate, err = parseDate(toDateStr)
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError(fmt.Sprintf("invalid to_date format: %v", err)))
			return
		}
	} else {
		// Default to now
		toDate = time.Now()
	}

	// Ensure from_date is before to_date
	if fromDate.After(toDate) {
		middleware.HandleError(w, r, errors.NewValidationError("from_date must be before to_date"))
		return
	}

	// Set time to start/end of day
	fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, fromDate.Location())
	toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, toDate.Location())

	// Execute query
	dashboardQuery := query.GetAdminDashboard{
		FromDate: fromDate,
		ToDate:   toDate,
	}

	result, err := c.dashboardHandler.Handle(r.Context(), dashboardQuery)
	if err != nil {
		middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("failed to get dashboard stats: %v", err)))
		return
	}

	// Format response with sorted data using custom JSON encoding
	responseData := map[string]interface{}{
		"new_user_registrations": convertToSortedJSON(result.NewUserRegistrations),
		"total_revenue":          convertToSortedJSONFloat(result.TotalRevenue),
		"scheduled_bookings":     convertToSortedJSON(result.ScheduledBookings),
		"summary": map[string]interface{}{
			"total_new_users": result.Summary.TotalNewUsers,
			"total_revenue":   result.Summary.TotalRevenue,
			"total_bookings":  result.Summary.TotalBookings,
			"from_date":       result.Summary.FromDate.Format("2006-01-02"),
			"to_date":         result.Summary.ToDate.Format("2006-01-02"),
		},
	}

	response.SendSuccess(w, r, responseData)
}

// convertToSortedJSON converts hierarchical int data to []byte with sorted keys
func convertToSortedJSON(data map[int]map[int]map[int]int) json.RawMessage {
	if data == nil || len(data) == 0 {
		return json.RawMessage("{}")
	}
	
	// Get sorted years
	years := make([]int, 0, len(data))
	for year := range data {
		years = append(years, year)
	}
	sort.Ints(years)
	
	// Build JSON manually with sorted keys
	result := "{"
	for i, year := range years {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%d":{`, year)
		
		// Get sorted months
		months := make([]int, 0, len(data[year]))
		for month := range data[year] {
			months = append(months, month)
		}
		sort.Ints(months)
		
		for j, month := range months {
			if j > 0 {
				result += ","
			}
			result += fmt.Sprintf(`"%d":{`, month)
			
			// Get sorted days
			days := make([]int, 0, len(data[year][month]))
			for day := range data[year][month] {
				days = append(days, day)
			}
			sort.Ints(days)
			
			for k, day := range days {
				if k > 0 {
					result += ","
				}
				result += fmt.Sprintf(`"%d":%d`, day, data[year][month][day])
			}
			result += "}"
		}
		result += "}"
	}
	result += "}"
	
	return json.RawMessage(result)
}

// convertToSortedJSONFloat converts hierarchical float data to []byte with sorted keys
func convertToSortedJSONFloat(data map[int]map[int]map[int]float64) json.RawMessage {
	if data == nil || len(data) == 0 {
		return json.RawMessage("{}")
	}
	
	// Get sorted years
	years := make([]int, 0, len(data))
	for year := range data {
		years = append(years, year)
	}
	sort.Ints(years)
	
	// Build JSON manually with sorted keys
	result := "{"
	for i, year := range years {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%d":{`, year)
		
		// Get sorted months
		months := make([]int, 0, len(data[year]))
		for month := range data[year] {
			months = append(months, month)
		}
		sort.Ints(months)
		
		for j, month := range months {
			if j > 0 {
				result += ","
			}
			result += fmt.Sprintf(`"%d":{`, month)
			
			// Get sorted days
			days := make([]int, 0, len(data[year][month]))
			for day := range data[year][month] {
				days = append(days, day)
			}
			sort.Ints(days)
			
			for k, day := range days {
				if k > 0 {
					result += ","
				}
				result += fmt.Sprintf(`"%d":%.2f`, day, data[year][month][day])
			}
			result += "}"
		}
		result += "}"
	}
	result += "}"
	
	return json.RawMessage(result)
}

// parseDate tries to parse date in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	// Try common formats
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		"2006-01-02 15:04:05",  // YYYY-MM-DD HH:MM:SS
		"01/02/2006",           // MM/DD/YYYY
		"02-01-2006",           // DD-MM-YYYY
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}

	return time.Time{}, lastErr
}
