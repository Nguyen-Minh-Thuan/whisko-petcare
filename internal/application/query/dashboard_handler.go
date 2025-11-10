package query

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminDashboardHandler struct {
	userCollection     *mongo.Collection
	paymentCollection  *mongo.Collection
	scheduleCollection *mongo.Collection
	serviceCollection  *mongo.Collection  // ← NEW: For fetching service names
}

func NewAdminDashboardHandler(db *mongo.Database) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		userCollection:     db.Collection("users"),
		paymentCollection:  db.Collection("payments"),
		scheduleCollection: db.Collection("schedules"),
		serviceCollection:  db.Collection("services"),  // ← NEW
	}
}

func (h *AdminDashboardHandler) Handle(ctx context.Context, query GetAdminDashboard) (*AdminDashboardResult, error) {
	fmt.Printf("📊 Dashboard Query: from %v to %v\n", query.FromDate.Format("2006-01-02"), query.ToDate.Format("2006-01-02"))
	
	// Initialize result
	result := &AdminDashboardResult{
		NewUserRegistrations: make(map[int]map[int]map[int]int),
		TotalRevenue:         make(map[int]map[int]map[int]float64),
		ScheduledBookings:    make(map[int]map[int]map[int]int),
		Summary: DashboardSummary{
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
		},
	}

	// Get new user registrations
	if err := h.getNewUserRegistrations(ctx, query, result); err != nil {
		return nil, fmt.Errorf("failed to get user registrations: %w", err)
	}
	fmt.Printf("✅ Users found: %d\n", result.Summary.TotalNewUsers)

	// Get revenue data
	if err := h.getTotalRevenue(ctx, query, result); err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}
	fmt.Printf("✅ Revenue: %.2f\n", result.Summary.TotalRevenue)

	// Get scheduled bookings
	if err := h.getScheduledBookings(ctx, query, result); err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}
	fmt.Printf("✅ Bookings: %d\n", result.Summary.TotalBookings)

	return result, nil
}

func (h *AdminDashboardHandler) getNewUserRegistrations(ctx context.Context, query GetAdminDashboard, result *AdminDashboardResult) error {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": query.FromDate,
			"$lte": query.ToDate,
		},
		"is_deleted": bson.M{"$ne": true}, // Exclude deleted users
	}

	cursor, err := h.userCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user bson.M
		if err := cursor.Decode(&user); err != nil {
			fmt.Printf("⚠️  Failed to decode user: %v\n", err)
			continue
		}

		// Handle created_at field - MongoDB stores it as primitive.DateTime which is converted to time.Time
		var createdAt time.Time
		switch v := user["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			fmt.Printf("⚠️  User %v has invalid created_at type: %T\n", user["_id"], user["created_at"])
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		// Initialize nested maps if needed
		if result.NewUserRegistrations[year] == nil {
			result.NewUserRegistrations[year] = make(map[int]map[int]int)
		}
		if result.NewUserRegistrations[year][month] == nil {
			result.NewUserRegistrations[year][month] = make(map[int]int)
		}

		// Increment count
		result.NewUserRegistrations[year][month][day]++
		result.Summary.TotalNewUsers++
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

func (h *AdminDashboardHandler) getTotalRevenue(ctx context.Context, query GetAdminDashboard, result *AdminDashboardResult) error {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": query.FromDate,
			"$lte": query.ToDate,
		},
		"status": "PAID", // Only count successful payments (uppercase)
	}

	cursor, err := h.paymentCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var payment bson.M
		if err := cursor.Decode(&payment); err != nil {
			fmt.Printf("⚠️  Failed to decode payment: %v\n", err)
			continue
		}

		// Handle created_at field
		var createdAt time.Time
		switch v := payment["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			fmt.Printf("⚠️  Payment %v has invalid created_at type: %T\n", payment["_id"], payment["created_at"])
			continue
		}

		// Get amount - handle both int and float
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
			fmt.Printf("⚠️  Payment %v has invalid amount type: %T\n", payment["_id"], payment["amount"])
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		// Initialize nested maps if needed
		if result.TotalRevenue[year] == nil {
			result.TotalRevenue[year] = make(map[int]map[int]float64)
		}
		if result.TotalRevenue[year][month] == nil {
			result.TotalRevenue[year][month] = make(map[int]float64)
		}

		// Add revenue
		result.TotalRevenue[year][month][day] += amount
		result.Summary.TotalRevenue += amount
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

func (h *AdminDashboardHandler) getScheduledBookings(ctx context.Context, query GetAdminDashboard, result *AdminDashboardResult) error {
	filter := bson.M{
		"created_at": bson.M{
			"$gte": query.FromDate,
			"$lte": query.ToDate,
		},
	}

	cursor, err := h.scheduleCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to query schedules: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var schedule bson.M
		if err := cursor.Decode(&schedule); err != nil {
			fmt.Printf("⚠️  Failed to decode schedule: %v\n", err)
			continue
		}

		// Handle created_at field
		var createdAt time.Time
		switch v := schedule["created_at"].(type) {
		case time.Time:
			createdAt = v
		case primitive.DateTime:
			createdAt = v.Time()
		default:
			fmt.Printf("⚠️  Schedule %v has invalid created_at type: %T\n", schedule["_id"], schedule["created_at"])
			continue
		}

		year := createdAt.Year()
		month := int(createdAt.Month())
		day := createdAt.Day()

		// Initialize nested maps if needed
		if result.ScheduledBookings[year] == nil {
			result.ScheduledBookings[year] = make(map[int]map[int]int)
		}
		if result.ScheduledBookings[year][month] == nil {
			result.ScheduledBookings[year][month] = make(map[int]int)
		}

		// Increment count
		result.ScheduledBookings[year][month][day]++
		result.Summary.TotalBookings++
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}
