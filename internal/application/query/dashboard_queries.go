package query

import "time"

// GetAdminDashboard query for admin dashboard statistics
type GetAdminDashboard struct {
	FromDate time.Time
	ToDate   time.Time
}

// AdminDashboardResult contains dashboard statistics organized by year/month/day
type AdminDashboardResult struct {
	NewUserRegistrations map[int]map[int]map[int]int       // Year -> Month -> Day -> Count
	TotalRevenue         map[int]map[int]map[int]float64   // Year -> Month -> Day -> Revenue
	ScheduledBookings    map[int]map[int]map[int]int       // Year -> Month -> Day -> Count
	Summary              DashboardSummary
}

// DashboardSummary contains aggregated totals
type DashboardSummary struct {
	TotalNewUsers      int
	TotalRevenue       float64
	TotalBookings      int
	FromDate           time.Time
	ToDate             time.Time
}

// GetVendorRevenue query for admin to view specific vendor's revenue
type GetVendorRevenue struct {
	VendorID string
	FromDate time.Time
	ToDate   time.Time
}

// VendorRevenueResult contains vendor revenue statistics
type VendorRevenueResult struct {
	VendorID          string
	TotalRevenue      map[int]map[int]map[int]float64   // Year -> Month -> Day -> Revenue
	ScheduledBookings map[int]map[int]map[int]int       // Year -> Month -> Day -> Count
	Summary           VendorRevenueSummary
}

// VendorRevenueSummary contains aggregated vendor totals
type VendorRevenueSummary struct {
	VendorID      string
	TotalRevenue  float64
	TotalBookings int
	FromDate      time.Time
	ToDate        time.Time
}

// GetVendorDashboard query for vendor to view their own dashboard
type GetVendorDashboard struct {
	VendorID string
	FromDate time.Time
	ToDate   time.Time
}

// VendorDashboardResult contains vendor's own dashboard statistics
type VendorDashboardResult struct {
	VendorID             string
	TotalRevenue         map[int]map[int]map[int]float64            // Year -> Month -> Day -> Total Revenue
	RevenueByService     map[string]map[int]map[int]map[int]float64 // ServiceID -> Year -> Month -> Day -> Revenue
	ScheduledBookings    map[int]map[int]map[int]int                // Year -> Month -> Day -> Count
	Summary              VendorDashboardSummary
	ServiceNames         map[string]string                           // ServiceID -> Service Name
}

// VendorDashboardSummary contains aggregated vendor dashboard totals
type VendorDashboardSummary struct {
	VendorID              string
	TotalRevenue          float64
	TotalBookings         int
	RevenueByServiceTotal map[string]float64 // ServiceID -> Total Revenue
	FromDate              time.Time
	ToDate                time.Time
}
