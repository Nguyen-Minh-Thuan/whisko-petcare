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

// PaymentItemReadModel represents an item in a payment
type PaymentItemReadModel struct {
	Name     string `json:"name" bson:"name"`
	Quantity int    `json:"quantity" bson:"quantity"`
	Price    int    `json:"price" bson:"price"`
}

// PaymentReadModel represents the read model for payments
type PaymentReadModel struct {
	ID                 string                  `json:"id" bson:"_id"`
	OrderCode          int64                   `json:"order_code" bson:"order_code"`
	UserID             string                  `json:"user_id" bson:"user_id"`
	Amount             int                     `json:"amount" bson:"amount"`
	Description        string                  `json:"description" bson:"description"`
	Items              []PaymentItemReadModel  `json:"items" bson:"items"`
	Status             string                  `json:"status" bson:"status"`
	Method             string                  `json:"method" bson:"method"`
	PayOSTransactionID string                  `json:"payos_transaction_id" bson:"payos_transaction_id"`
	CheckoutURL        string                  `json:"checkout_url" bson:"checkout_url"`
	QRCode             string                  `json:"qr_code" bson:"qr_code"`
	ExpiredAt          time.Time               `json:"expired_at" bson:"expired_at"`
	Version            int                     `json:"version" bson:"version"`
	CreatedAt          time.Time               `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at" bson:"updated_at"`
}

// PaymentProjection defines operations for payment read model
type PaymentProjection interface {
	GetByID(ctx context.Context, id string) (*PaymentReadModel, error)
	GetByOrderCode(ctx context.Context, orderCode int64) (*PaymentReadModel, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*PaymentReadModel, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*PaymentReadModel, error)

	// Event handlers
	HandlePaymentCreated(ctx context.Context, event *event.PaymentCreated) error
	HandlePaymentUpdated(ctx context.Context, event *event.PaymentUpdated) error
	HandlePaymentStatusChanged(ctx context.Context, event *event.PaymentStatusChanged) error
}

// MongoPaymentProjection implements PaymentProjection using MongoDB
type MongoPaymentProjection struct {
	collection *mongo.Collection
}

func NewMongoPaymentProjection(database *mongo.Database) PaymentProjection {
	return &MongoPaymentProjection{
		collection: database.Collection("payments_read"),
	}
}

func (p *MongoPaymentProjection) GetByID(ctx context.Context, id string) (*PaymentReadModel, error) {
	var payment PaymentReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

func (p *MongoPaymentProjection) GetByOrderCode(ctx context.Context, orderCode int64) (*PaymentReadModel, error) {
	var payment PaymentReadModel
	err := p.collection.FindOne(ctx, bson.M{"order_code": orderCode}).Decode(&payment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

func (p *MongoPaymentProjection) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*PaymentReadModel, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := p.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*PaymentReadModel
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	return payments, nil
}

func (p *MongoPaymentProjection) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*PaymentReadModel, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := p.collection.Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*PaymentReadModel
	if err := cursor.All(ctx, &payments); err != nil {
		return nil, fmt.Errorf("failed to decode payments: %w", err)
	}

	return payments, nil
}

// Event handlers
func (p *MongoPaymentProjection) HandlePaymentCreated(ctx context.Context, evt *event.PaymentCreated) error {
	fmt.Printf("DEBUG: HandlePaymentCreated called for payment %s\n", evt.PaymentID)
	
	items := make([]PaymentItemReadModel, len(evt.Items))
	for i, item := range evt.Items {
		items[i] = PaymentItemReadModel{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	payment := PaymentReadModel{
		ID:          evt.PaymentID,
		OrderCode:   evt.OrderCode,
		UserID:      evt.UserID,
		Amount:      evt.Amount,
		Description: evt.Description,
		Items:       items,
		Status:      evt.Status,
		Method:      evt.Method,
		ExpiredAt:   evt.ExpiredAt,
		Version:     1,
		CreatedAt:   evt.Timestamp,
		UpdatedAt:   evt.Timestamp,
	}

	fmt.Printf("DEBUG: Inserting payment into payments_read collection\n")
	_, err := p.collection.InsertOne(ctx, payment)
	if err != nil {
		fmt.Printf("ERROR: Failed to insert payment to read model: %v\n", err)
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	fmt.Printf("DEBUG: Successfully inserted payment %s to read model\n", evt.PaymentID)
	return nil
}

func (p *MongoPaymentProjection) HandlePaymentUpdated(ctx context.Context, evt *event.PaymentUpdated) error {
	update := bson.M{
		"$set": bson.M{
			"payos_transaction_id": evt.PayOSTransactionID,
			"checkout_url":         evt.CheckoutURL,
			"qr_code":              evt.QRCode,
			"updated_at":           evt.Timestamp,
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	result, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.PaymentID}, update)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

func (p *MongoPaymentProjection) HandlePaymentStatusChanged(ctx context.Context, evt *event.PaymentStatusChanged) error {
	update := bson.M{
		"$set": bson.M{
			"status":     evt.NewStatus,
			"updated_at": evt.Timestamp,
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	result, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.PaymentID}, update)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}
