package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

type HTTPVendorDashboardController struct {
	dashboardHandler *query.AdminDashboardHandler
}

func NewHTTPVendorDashboardController(dashboardHandler *query.AdminDashboardHandler) *HTTPVendorDashboardController {
	return &HTTPVendorDashboardController{
		dashboardHandler: dashboardHandler,
	}
}

// GetVendorRevenue handles GET /admin/vendors/{vendorID}/revenue (Admin only)
func (c *HTTPVendorDashboardController) GetVendorRevenue(w http.ResponseWriter, r *http.Request) {
	// Extract vendor ID from path
	vendorID := extractVendorIDFromPath(r.URL.Path, "/admin/vendors/", "/revenue")
	if vendorID == "" {
		middleware.HandleError(w, r, errors.NewValidationError("vendor ID is required"))
		return
	}

	// Parse date range
	fromDate, toDate, err := parseDateRange(r)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Execute query
	vendorQuery := query.GetVendorRevenue{
		VendorID: vendorID,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	result, err := c.dashboardHandler.HandleVendorRevenue(r.Context(), vendorQuery)
	if err != nil {
		middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("failed to get vendor revenue: %v", err)))
		return
	}

	// Format response with sorted data
	responseData := map[string]interface{}{
		"vendor_id":          result.VendorID,
		"total_revenue":      convertToSortedJSONFloat(result.TotalRevenue),
		"scheduled_bookings": convertToSortedJSON(result.ScheduledBookings),
		"summary": map[string]interface{}{
			"vendor_id":      result.Summary.VendorID,
			"total_revenue":  result.Summary.TotalRevenue,
			"total_bookings": result.Summary.TotalBookings,
			"from_date":      result.Summary.FromDate.Format("2006-01-02"),
			"to_date":        result.Summary.ToDate.Format("2006-01-02"),
		},
	}

	response.SendSuccess(w, r, responseData)
}

// GetVendorDashboardByAdmin handles GET /admin/vendors/{vendorID}/dashboard (Admin only)
// This allows admins to view any vendor's dashboard without needing the vendor's credentials
func (c *HTTPVendorDashboardController) GetVendorDashboardByAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔐 Admin accessing vendor dashboard")
	
	// Extract vendor ID from path
	vendorID := extractVendorIDFromPath(r.URL.Path, "/admin/vendors/", "/dashboard")
	if vendorID == "" {
		fmt.Println("   ❌ No vendor ID provided in path")
		middleware.HandleError(w, r, errors.NewValidationError("vendor ID is required"))
		return
	}
	
	fmt.Printf("   Requested vendor ID: %s\n", vendorID)

	// Parse date range
	fromDate, toDate, err := parseDateRange(r)
	if err != nil {
		fmt.Printf("   ❌ Date parse error: %v\n", err)
		middleware.HandleError(w, r, err)
		return
	}
	
	fmt.Printf("   Date range: %s to %s\n", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	// Execute query
	vendorQuery := query.GetVendorDashboard{
		VendorID: vendorID,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	result, err := c.dashboardHandler.HandleVendorDashboard(r.Context(), vendorQuery)
	if err != nil {
		fmt.Printf("   ❌ Handler error: %v\n", err)
		middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("failed to get vendor dashboard: %v", err)))
		return
	}

	fmt.Printf("   ✅ Dashboard retrieved: Revenue=$%.2f, Bookings=%d\n", 
		result.Summary.TotalRevenue, result.Summary.TotalBookings)

	// Format response with sorted data
	responseData := map[string]interface{}{
		"vendor_id":          result.VendorID,
		"total_revenue":      convertToSortedJSONFloat(result.TotalRevenue),
		"revenue_by_service": convertRevenueByServiceToJSON(result.RevenueByService),
		"service_names":      result.ServiceNames,  // ← NEW: Include service names
		"scheduled_bookings": convertToSortedJSON(result.ScheduledBookings),
		"summary": map[string]interface{}{
			"vendor_id":                result.Summary.VendorID,
			"total_revenue":            result.Summary.TotalRevenue,
			"total_bookings":           result.Summary.TotalBookings,
			"revenue_by_service_total": result.Summary.RevenueByServiceTotal,
			"from_date":                result.Summary.FromDate.Format("2006-01-02"),
			"to_date":                  result.Summary.ToDate.Format("2006-01-02"),
		},
	}

	response.SendSuccess(w, r, responseData)
}

// GetVendorDashboard handles GET /vendors/dashboard (Vendor only - their own data)
func (c *HTTPVendorDashboardController) GetVendorDashboard(w http.ResponseWriter, r *http.Request) {
	// Get vendor ID from JWT context (set by middleware)
	vendorID := r.Context().Value("vendor_id")
	if vendorID == nil {
		// Fallback: try to get from URL parameter
		vendorID = r.URL.Query().Get("vendor_id")
		if vendorID == "" {
			middleware.HandleError(w, r, errors.NewForbiddenError("vendor ID not found in context"))
			return
		}
	}

	// Parse date range
	fromDate, toDate, err := parseDateRange(r)
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Execute query
	vendorQuery := query.GetVendorDashboard{
		VendorID: vendorID.(string),
		FromDate: fromDate,
		ToDate:   toDate,
	}

	result, err := c.dashboardHandler.HandleVendorDashboard(r.Context(), vendorQuery)
	if err != nil {
		middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("failed to get vendor dashboard: %v", err)))
		return
	}

	// Format response with sorted data
	responseData := map[string]interface{}{
		"vendor_id":          result.VendorID,
		"total_revenue":      convertToSortedJSONFloat(result.TotalRevenue),
		"revenue_by_service": convertRevenueByServiceToJSON(result.RevenueByService),
		"service_names":      result.ServiceNames,  // ← NEW: Include service names
		"scheduled_bookings": convertToSortedJSON(result.ScheduledBookings),
		"summary": map[string]interface{}{
			"vendor_id":                result.Summary.VendorID,
			"total_revenue":            result.Summary.TotalRevenue,
			"total_bookings":           result.Summary.TotalBookings,
			"revenue_by_service_total": result.Summary.RevenueByServiceTotal,
			"from_date":                result.Summary.FromDate.Format("2006-01-02"),
			"to_date":                  result.Summary.ToDate.Format("2006-01-02"),
		},
	}

	response.SendSuccess(w, r, responseData)
}

// Helper: Parse date range from query parameters
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate time.Time
	var err error

	if fromDateStr != "" {
		fromDate, err = parseDate(fromDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.NewValidationError(fmt.Sprintf("invalid from_date format: %v", err))
		}
	} else {
		fromDate = time.Now().AddDate(0, 0, -30)
	}

	if toDateStr != "" {
		toDate, err = parseDate(toDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.NewValidationError(fmt.Sprintf("invalid to_date format: %v", err))
		}
	} else {
		toDate = time.Now()
	}

	if fromDate.After(toDate) {
		return time.Time{}, time.Time{}, errors.NewValidationError("from_date must be before to_date")
	}

	// Set time to start/end of day
	fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, fromDate.Location())
	toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, toDate.Location())

	return fromDate, toDate, nil
}

// Helper: Extract vendor ID from URL path
func extractVendorIDFromPath(path, prefix, suffix string) string {
	// Remove prefix
	path = strings.TrimPrefix(path, prefix)
	// Remove suffix
	path = strings.TrimSuffix(path, suffix)
	return path
}

// Helper: Convert revenue by service to sorted JSON
func convertRevenueByServiceToJSON(data map[string]map[int]map[int]map[int]float64) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage)
	
	for serviceID, serviceData := range data {
		result[serviceID] = convertToSortedJSONFloat(serviceData)
	}
	
	return result
}
