package query

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HandleVendorRevenue handles admin query for specific vendor's revenue
func (h *AdminDashboardHandler) HandleVendorRevenue(ctx context.Context, query GetVendorRevenue) (*VendorRevenueResult, error) {
	fmt.Printf("📊 Vendor Revenue Query: VendorID=%s, from %v to %v\n", 
		query.VendorID, query.FromDate.Format("2006-01-02"), query.ToDate.Format("2006-01-02"))
	
	result := &VendorRevenueResult{
		VendorID:          query.VendorID,
		TotalRevenue:      make(map[int]map[int]map[int]float64),
		ScheduledBookings: make(map[int]map[int]map[int]int),
		Summary: VendorRevenueSummary{
			VendorID: query.VendorID,
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
		},
	}

	// Get revenue for this vendor
	if err := h.getVendorRevenue(ctx, query.VendorID, query.FromDate, query.ToDate, result); err != nil {
		return nil, fmt.Errorf("failed to get vendor revenue: %w", err)
	}

	// Get bookings for this vendor
	if err := h.getVendorBookings(ctx, query.VendorID, query.FromDate, query.ToDate, result); err != nil {
		return nil, fmt.Errorf("failed to get vendor bookings: %w", err)
	}

	fmt.Printf("✅ Vendor Revenue: %.2f, Bookings: %d\n", result.Summary.TotalRevenue, result.Summary.TotalBookings)
	return result, nil
}

// HandleVendorDashboard handles vendor's own dashboard query
func (h *AdminDashboardHandler) HandleVendorDashboard(ctx context.Context, query GetVendorDashboard) (*VendorDashboardResult, error) {
	fmt.Printf("📊 Vendor Dashboard Query: VendorID=%s, from %v to %v\n", 
		query.VendorID, query.FromDate.Format("2006-01-02"), query.ToDate.Format("2006-01-02"))
	
	result := &VendorDashboardResult{
		VendorID:          query.VendorID,
		TotalRevenue:      make(map[int]map[int]map[int]float64),
		RevenueByService:  make(map[string]map[int]map[int]map[int]float64),
		ScheduledBookings: make(map[int]map[int]map[int]int),
		ServiceNames:      make(map[string]string),  // ← NEW: Initialize service names map
		Summary: VendorDashboardSummary{
			VendorID:              query.VendorID,
			RevenueByServiceTotal: make(map[string]float64),
			FromDate:              query.FromDate,
			ToDate:                query.ToDate,
		},
	}

	// Get total revenue and revenue by service
	if err := h.getVendorDashboardRevenue(ctx, query.VendorID, query.FromDate, query.ToDate, result); err != nil {
		return nil, fmt.Errorf("failed to get vendor dashboard revenue: %w", err)
	}

	// Get bookings
	if err := h.getVendorDashboardBookings(ctx, query.VendorID, query.FromDate, query.ToDate, result); err != nil {
		return nil, fmt.Errorf("failed to get vendor dashboard bookings: %w", err)
	}

	fmt.Printf("✅ Vendor Dashboard - Revenue: %.2f, Bookings: %d, Services: %d\n", 
		result.Summary.TotalRevenue, result.Summary.TotalBookings, len(result.Summary.RevenueByServiceTotal))
	return result, nil
}

// Helper: Get vendor revenue from payments
func (h *AdminDashboardHandler) getVendorRevenue(ctx context.Context, vendorID string, fromDate, toDate time.Time, 
	result *VendorRevenueResult) error {
	
	fmt.Printf("🔍 getVendorRevenue: Searching for vendor %s\n", vendorID)
	
	// FIXED: Query payments directly by vendor_id (payments don't have schedule_id)
	filter := bson.M{
		"created_at": bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		},
		"status": "PAID",
		"vendor_id": vendorID,  // ← FIXED: Direct vendor filter
	}
	
	fmt.Printf("   Payment filter: %+v\n", filter)

	cursor, err := h.paymentCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}
	defer cursor.Close(ctx)

	paymentCount := 0
	
	for cursor.Next(ctx) {
		paymentCount++
		var payment bson.M
		if err := cursor.Decode(&payment); err != nil {
			fmt.Printf("   ⚠️  Failed to decode payment: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Payment matched! ID=%v, Amount=%v\n", payment["_id"], payment["amount"])

		// Process payment
		var createdAt time.Time
		switch v := payment["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			continue
		}

		var amount float64
		switch v := payment["amount"].(type) {
		case int:
			amount = float64(v)
		case int32:
			amount = float64(v)
		case int64:
			amount = float64(v)
		case float32:
			amount = float64(v)
		case float64:
			amount = v
		default:
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		if result.TotalRevenue[year] == nil {
			result.TotalRevenue[year] = make(map[int]map[int]float64)
		}
		if result.TotalRevenue[year][month] == nil {
			result.TotalRevenue[year][month] = make(map[int]float64)
		}

		result.TotalRevenue[year][month][day] += amount
		result.Summary.TotalRevenue += amount
	}

	fmt.Printf("📊 Vendor Revenue Summary: %d payments matched vendor %s, Total: $%.2f\n",
		paymentCount, vendorID, result.Summary.TotalRevenue)

	return cursor.Err()
}

// Helper: Get vendor bookings
func (h *AdminDashboardHandler) getVendorBookings(ctx context.Context, vendorID string, fromDate, toDate time.Time,
	result *VendorRevenueResult) error {
	
	fmt.Printf("🔍 getVendorBookings: Searching for vendor %s\n", vendorID)
	
	// FIXED: Schedules have shop_id nested inside booked_shop object
	filter := bson.M{
		"booked_shop.shop_id": vendorID,  // ← FIXED: Use nested field path
		"created_at": bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		},
	}
	
	fmt.Printf("   Schedule filter: %+v\n", filter)

	cursor, err := h.scheduleCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query schedules: %w", err)
	}
	defer cursor.Close(ctx)

	bookingCount := 0
	
	for cursor.Next(ctx) {
		bookingCount++
		var schedule bson.M
		if err := cursor.Decode(&schedule); err != nil {
			fmt.Printf("   ⚠️  Failed to decode schedule: %v\n", err)
			continue
		}

		var createdAt time.Time
		switch v := schedule["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		if result.ScheduledBookings[year] == nil {
			result.ScheduledBookings[year] = make(map[int]map[int]int)
		}
		if result.ScheduledBookings[year][month] == nil {
			result.ScheduledBookings[year][month] = make(map[int]int)
		}

		result.ScheduledBookings[year][month][day]++
		result.Summary.TotalBookings++
	}

	fmt.Printf("📊 Vendor Bookings Summary: %d bookings found for vendor %s\n", 
		bookingCount, vendorID)

	return cursor.Err()
}

// Helper: Get vendor dashboard revenue (total + by service)
func (h *AdminDashboardHandler) getVendorDashboardRevenue(ctx context.Context, vendorID string, fromDate, toDate time.Time,
	result *VendorDashboardResult) error {
	
	fmt.Printf("🔍 getVendorDashboardRevenue: Searching for vendor %s\n", vendorID)
	
	// FIXED: Query payments directly by vendor_id (payments don't have schedule_id)
	filter := bson.M{
		"created_at": bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		},
		"status": "PAID",
		"vendor_id": vendorID,  // ← FIXED: Direct vendor filter
	}
	
	fmt.Printf("   Payment filter: %+v\n", filter)

	cursor, err := h.paymentCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}
	defer cursor.Close(ctx)

	paymentCount := 0

	for cursor.Next(ctx) {
		paymentCount++
		var payment bson.M
		if err := cursor.Decode(&payment); err != nil {
			fmt.Printf("   ⚠️  Failed to decode payment: %v\n", err)
			continue
		}

		// Get service IDs from payment (payments store service_ids directly)
		var serviceIDs []string
		if sids, ok := payment["service_ids"].(primitive.A); ok {
			for _, sid := range sids {
				if sidStr, ok := sid.(string); ok {
					serviceIDs = append(serviceIDs, sidStr)
				}
			}
		}
		
		// Use first service ID or "unknown"
		serviceID := "unknown"
		if len(serviceIDs) > 0 {
			serviceID = serviceIDs[0]
		}
		
		fmt.Printf("   ✅ Payment matched! Service: %s, Amount: %v\n", serviceID, payment["amount"])

		// Process payment
		var createdAt time.Time
		switch v := payment["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			continue
		}

		var amount float64
		switch v := payment["amount"].(type) {
		case int:
			amount = float64(v)
		case int32:
			amount = float64(v)
		case int64:
			amount = float64(v)
		case float32:
			amount = float64(v)
		case float64:
			amount = v
		default:
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		// Add to total revenue
		if result.TotalRevenue[year] == nil {
			result.TotalRevenue[year] = make(map[int]map[int]float64)
		}
		if result.TotalRevenue[year][month] == nil {
			result.TotalRevenue[year][month] = make(map[int]float64)
		}
		result.TotalRevenue[year][month][day] += amount
		result.Summary.TotalRevenue += amount

		// Add to service revenue
		if result.RevenueByService[serviceID] == nil {
			result.RevenueByService[serviceID] = make(map[int]map[int]map[int]float64)
		}
		if result.RevenueByService[serviceID][year] == nil {
			result.RevenueByService[serviceID][year] = make(map[int]map[int]float64)
		}
		if result.RevenueByService[serviceID][year][month] == nil {
			result.RevenueByService[serviceID][year][month] = make(map[int]float64)
		}
		result.RevenueByService[serviceID][year][month][day] += amount
		result.Summary.RevenueByServiceTotal[serviceID] += amount
		
		// Fetch service name if not already cached
		if _, exists := result.ServiceNames[serviceID]; !exists && serviceID != "unknown" {
			var service bson.M
			err := h.serviceCollection.FindOne(ctx, bson.M{"_id": serviceID}).Decode(&service)
			if err == nil {
				if name, ok := service["name"].(string); ok {
					result.ServiceNames[serviceID] = name
					fmt.Printf("      📝 Cached service name: %s -> %s\n", serviceID, name)
				}
			} else {
				result.ServiceNames[serviceID] = "Unknown Service"
			}
		}
	}

	fmt.Printf("📊 Vendor Dashboard Revenue: %d payments matched vendor %s, Total: $%.2f\n",
		paymentCount, vendorID, result.Summary.TotalRevenue)

	return cursor.Err()
}

// Helper: Get vendor dashboard bookings
func (h *AdminDashboardHandler) getVendorDashboardBookings(ctx context.Context, vendorID string, fromDate, toDate time.Time,
	result *VendorDashboardResult) error {
	
	fmt.Printf("🔍 getVendorDashboardBookings: Searching for vendor %s\n", vendorID)
	
	// FIXED: Schedules have shop_id nested inside booked_shop object
	filter := bson.M{
		"booked_shop.shop_id": vendorID,  // ← FIXED: Use nested field path
		"created_at": bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		},
	}
	
	fmt.Printf("   Schedule filter: %+v\n", filter)

	cursor, err := h.scheduleCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query schedules: %w", err)
	}
	defer cursor.Close(ctx)

	bookingCount := 0
	
	// First, let's check if there are ANY schedules in the collection
	totalCount, _ := h.scheduleCollection.CountDocuments(ctx, bson.M{})
	fmt.Printf("   📊 Total schedules in collection: %d\n", totalCount)
	
	// Check schedules in date range (any vendor)
	anyVendorSchedules, _ := h.scheduleCollection.CountDocuments(ctx, bson.M{
		"created_at": bson.M{"$gte": fromDate, "$lte": toDate},
	})
	fmt.Printf("   📊 Schedules in date range: %d\n", anyVendorSchedules)
	
	// Let's check a sample schedule to see its structure
	var sampleSchedule bson.M
	if err := h.scheduleCollection.FindOne(ctx, bson.M{}).Decode(&sampleSchedule); err == nil {
		fmt.Printf("   🔍 Sample schedule fields: ")
		for key := range sampleSchedule {
			fmt.Printf("%s, ", key)
		}
		fmt.Println()
		if shopID, ok := sampleSchedule["shop_id"]; ok {
			fmt.Printf("   🔍 Sample shop_id value: %v (type: %T)\n", shopID, shopID)
		}
		if bookedShop, ok := sampleSchedule["booked_shop"].(bson.M); ok {
			if shopID, ok := bookedShop["shop_id"]; ok {
				fmt.Printf("   🔍 Sample booked_shop.shop_id value: %v (type: %T)\n", shopID, shopID)
			}
		}
	}
	
	for cursor.Next(ctx) {
		bookingCount++
		var schedule bson.M
		if err := cursor.Decode(&schedule); err != nil {
			fmt.Printf("   ⚠️  Failed to decode schedule: %v\n", err)
			continue
		}

		var createdAt time.Time
		switch v := schedule["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		if result.ScheduledBookings[year] == nil {
			result.ScheduledBookings[year] = make(map[int]map[int]int)
		}
		if result.ScheduledBookings[year][month] == nil {
			result.ScheduledBookings[year][month] = make(map[int]int)
		}

		result.ScheduledBookings[year][month][day]++
		result.Summary.TotalBookings++
	}

	fmt.Printf("📊 Vendor Dashboard Bookings Summary: %d bookings found for vendor %s\n", 
		bookingCount, vendorID)

	return cursor.Err()
}
