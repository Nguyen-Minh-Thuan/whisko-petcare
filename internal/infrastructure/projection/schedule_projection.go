package projection

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ScheduleReadModel represents the read model for schedules
type ScheduleReadModel struct {
	ID           string              `bson:"_id" json:"id"`
	UserID       string              `bson:"user_id" json:"user_id"`
	ShopID       string              `bson:"shop_id" json:"shop_id"`
	PetID        string              `bson:"pet_id" json:"pet_id"`
	BookingUser  BookingUserRead     `bson:"booking_user" json:"booking_user"`
	BookedShop   BookedShopRead      `bson:"booked_shop" json:"booked_shop"`
	AssignedPet  AssignedPetRead     `bson:"assigned_pet" json:"assigned_pet"`
	StartTime    time.Time           `bson:"start_time" json:"start_time"`
	EndTime      time.Time           `bson:"end_time" json:"end_time"`
	Status       string              `bson:"status" json:"status"`
	IsActive     bool                `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time           `bson:"updated_at" json:"updated_at"`
}

type BookingUserRead struct {
	UserID  string `bson:"user_id" json:"user_id"`
	Name    string `bson:"name" json:"name"`
	Email   string `bson:"email" json:"email"`
	Phone   string `bson:"phone" json:"phone"`
	Address string `bson:"address" json:"address"`
}

type BookedShopRead struct {
	ShopID   string               `bson:"shop_id" json:"shop_id"`
	Name     string               `bson:"name" json:"name"`
	Location string               `bson:"location" json:"location"`
	Phone    string               `bson:"phone" json:"phone"`
	Services []BookedServiceRead  `bson:"booked_services" json:"services"`
}

type BookedServiceRead struct {
	ServiceID string `bson:"service_id" json:"service_id"`
	Name      string `bson:"name" json:"name"`
}

type AssignedPetRead struct {
	PetID   string  `bson:"pet_id" json:"pet_id"`
	Name    string  `bson:"name" json:"name"`
	Species string  `bson:"species" json:"species"`
	Breed   string  `bson:"breed" json:"breed"`
	Age     int     `bson:"age" json:"age"`
	Weight  float64 `bson:"weight" json:"weight"`
}

// MongoScheduleProjection implements ScheduleProjection using MongoDB
type MongoScheduleProjection struct {
	collection *mongo.Collection
}

// NewMongoScheduleProjection creates a new MongoDB schedule projection
func NewMongoScheduleProjection(db *mongo.Database) *MongoScheduleProjection {
	collection := db.Collection("schedules")
	
	// Create indexes
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "shop_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "pet_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "is_active", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "start_time", Value: 1},
			},
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		fmt.Printf("Warning: failed to create schedule indexes: %v\n", err)
	}
	
	return &MongoScheduleProjection{
		collection: collection,
	}
}

// GetByID retrieves a schedule by ID
func (p *MongoScheduleProjection) GetByID(ctx context.Context, id string) (interface{}, error) {
	var schedule ScheduleReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&schedule)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

// GetByUserID retrieves schedules by user ID with pagination
func (p *MongoScheduleProjection) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "start_time", Value: -1}})
	
	// Only set limit if it's greater than 0, otherwise return all
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	// Query both top-level user_id and nested booking_user.user_id for backward compatibility
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"booking_user.user_id": userID},
		},
		"is_active": true,
	}
	fmt.Printf("🔍 GetByUserID - Querying schedules with filter: %+v\n", filter)
	
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		fmt.Printf("❌ GetByUserID - Query error: %v\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var schedules []interface{}
	for cursor.Next(ctx) {
		var schedule ScheduleReadModel
		if err := cursor.Decode(&schedule); err != nil {
			fmt.Printf("❌ GetByUserID - Decode error: %v\n", err)
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	
	if err := cursor.Err(); err != nil {
		fmt.Printf("❌ GetByUserID - Cursor error: %v\n", err)
		return nil, err
	}
	
	fmt.Printf("✅ GetByUserID - Found %d schedules for user_id: %s\n", len(schedules), userID)
	return schedules, nil
}

// GetByShopID retrieves schedules by shop ID with pagination
func (p *MongoScheduleProjection) GetByShopID(ctx context.Context, shopID string, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "start_time", Value: -1}})
	
	// Only set limit if it's greater than 0, otherwise return all
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	// Query both top-level shop_id and nested booked_shop.shop_id for backward compatibility
	filter := bson.M{
		"$or": []bson.M{
			{"shop_id": shopID},
			{"booked_shop.shop_id": shopID},
		},
		"is_active": true,
	}
	fmt.Printf("🔍 GetByShopID - Querying schedules with filter: %+v\n", filter)
	
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		fmt.Printf("❌ GetByShopID - Query error: %v\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var schedules []interface{}
	for cursor.Next(ctx) {
		var schedule ScheduleReadModel
		if err := cursor.Decode(&schedule); err != nil {
			fmt.Printf("❌ GetByShopID - Decode error: %v\n", err)
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	
	if err := cursor.Err(); err != nil {
		fmt.Printf("❌ GetByShopID - Cursor error: %v\n", err)
		return nil, err
	}
	
	fmt.Printf("✅ GetByShopID - Found %d schedules for shop_id: %s\n", len(schedules), shopID)
	return schedules, nil
}

// ListAll retrieves all schedules with pagination
func (p *MongoScheduleProjection) ListAll(ctx context.Context, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "start_time", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var schedules []interface{}
	for cursor.Next(ctx) {
		var schedule ScheduleReadModel
		if err := cursor.Decode(&schedule); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return schedules, nil
}

// HandleScheduleCreated handles ScheduleCreated event
func (p *MongoScheduleProjection) HandleScheduleCreated(ctx context.Context, evt event.ScheduleCreated) error {
	schedule := ScheduleReadModel{
		ID:        evt.ScheduleID,
		UserID:    evt.BookingUser.UserID,
		ShopID:    evt.BookedVendor.ShopID,
		PetID:     evt.AssignedPet.PetID,
		BookingUser: BookingUserRead{
			UserID:  evt.BookingUser.UserID,
			Name:    evt.BookingUser.Name,
			Email:   evt.BookingUser.Email,
			Phone:   evt.BookingUser.Phone,
			Address: evt.BookingUser.Address,
		},
		BookedShop: BookedShopRead{
			ShopID:   evt.BookedVendor.ShopID,
			Name:     evt.BookedVendor.Name,
			Location: evt.BookedVendor.Location,
			Phone:    evt.BookedVendor.Phone,
			Services: convertToBookedServiceRead(evt.BookedVendor.BookedServices),
		},
		AssignedPet: AssignedPetRead{
			PetID:   evt.AssignedPet.PetID,
			Name:    evt.AssignedPet.Name,
			Species: evt.AssignedPet.Species,
			Breed:   evt.AssignedPet.Breed,
			Age:     evt.AssignedPet.Age,
			Weight:  evt.AssignedPet.Weight,
		},
		StartTime: evt.StartTime,
		EndTime:   evt.EndTime,
		Status:    evt.Status,
		IsActive:  true,
		CreatedAt: evt.Timestamp,
		UpdatedAt: evt.Timestamp,
	}
	
	fmt.Printf("📝 HandleScheduleCreated - Creating schedule with UserID: %s, ShopID: %s, PetID: %s\n", 
		evt.BookingUser.UserID, evt.BookedVendor.ShopID, evt.AssignedPet.PetID)
	
	_, err := p.collection.InsertOne(ctx, schedule)
	if err != nil {
		return fmt.Errorf("failed to insert schedule: %w", err)
	}
	
	return nil
}

// Helper function to convert BookedServicesData to BookedServiceRead
func convertToBookedServiceRead(services []event.BookedServicesData) []BookedServiceRead {
	result := make([]BookedServiceRead, len(services))
	for i, s := range services {
		result[i] = BookedServiceRead{
			ServiceID: s.ServiceID,
			Name:      s.Name,
		}
	}
	return result
}

// HandleScheduleStatusChanged handles ScheduleStatusChanged event
func (p *MongoScheduleProjection) HandleScheduleStatusChanged(ctx context.Context, evt event.ScheduleStatusChanged) error {
	update := bson.M{
		"$set": bson.M{
			"status":     evt.NewStatus,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.ScheduleID}, update)
	if err != nil {
		return fmt.Errorf("failed to update schedule status: %w", err)
	}
	
	return nil
}

// HandleScheduleCompleted handles ScheduleCompleted event
func (p *MongoScheduleProjection) HandleScheduleCompleted(ctx context.Context, evt event.ScheduleCompleted) error {
	update := bson.M{
		"$set": bson.M{
			"status":     "completed",
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.ScheduleID}, update)
	if err != nil {
		return fmt.Errorf("failed to complete schedule: %w", err)
	}
	
	return nil
}

// HandleScheduleCancelled handles ScheduleCancelled event
func (p *MongoScheduleProjection) HandleScheduleCancelled(ctx context.Context, evt event.ScheduleCancelled) error {
	update := bson.M{
		"$set": bson.M{
			"status":     "cancelled",
			"is_active":  false,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.ScheduleID}, update)
	if err != nil {
		return fmt.Errorf("failed to cancel schedule: %w", err)
	}
	
	return nil
	
}
