package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConfig holds configuration for MongoDB connection
type MongoConfig struct {
	URI      string
	Database string
	Username string
	Password string
	Timeout  time.Duration
}

// MongoClient wraps the MongoDB client and database
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoConfig
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(config *MongoConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Set client options
	clienOptions := options.Client().
		ApplyURI(config.URI).
		SetServerSelectionTimeout(config.Timeout)

	if config.Username != "" && config.Password != "" {
		clienOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clienOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &MongoClient{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// GetDatabase returns the MongoDB database
func (mc *MongoClient) GetDatabase() *mongo.Database {
	return mc.database
}

// GetClient returns the underlying MongoDB client
func (mc *MongoClient) GetClient() *mongo.Client {
	return mc.client
}

// GetCollection returns a MongoDB collection
func (mc *MongoClient) GetCollection(name string) *mongo.Collection {
	return mc.database.Collection(name)
}

// Close closes the MongoDB connection
func (mc *MongoClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), mc.config.Timeout)
	defer cancel()

	return mc.client.Disconnect(ctx)
}

// Ping tests the MongoDB connection
func (mc *MongoClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), mc.config.Timeout)
	defer cancel()

	return mc.client.Ping(ctx, nil)
}
