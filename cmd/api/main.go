package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/infrastructure/bus"
	httpHandler "whisko-petcare/internal/infrastructure/http"
	"whisko-petcare/internal/infrastructure/mongo"
	"whisko-petcare/internal/infrastructure/projection"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	log.Println("Starting Whisko Pet Care API (Event Sourcing)...")

	mongoConfig := &mongo.MongoConfig{
		URI:      getEnv("MONGO_URI", ""),
		Database: getEnv("MONGO_DATABASE", ""),
		Timeout:  30 * time.Second,
	}

	// Initialize MongoDB client
	mongoClient, err := mongo.NewMongoClient(mongoConfig)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			log.Printf("Error closing MongoDB connection: %v", err)
		}
	}()

	// Test MongoDB connection
	if err := mongoClient.Ping(); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("✅ Connected to MongoDB successfully")

	// Initialize infrastructure
	database := mongoClient.GetDatabase()
	eventBus := bus.NewInMemoryEventBus()
	userProjection := projection.NewMongoUserProjection(database)

	// Initialize Unit of Work factory
	uowFactory := mongo.NewMongoUnitOfWorkFactory(mongoClient.GetClient(), database)

	// Subscribe projection to events
	eventBus.Subscribe("UserCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserCreated(ctx, e.(*event.UserCreated))
		}))

	eventBus.Subscribe("UserProfileUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserProfileUpdated(ctx, e.(*event.UserProfileUpdated))
		}))

	eventBus.Subscribe("UserContactUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserContactUpdated(ctx, e.(*event.UserContactUpdated))
		}))

	eventBus.Subscribe("UserDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserDeleted(ctx, e.(*event.UserDeleted))
		}))

	// Initialize Unit of Work command handlers
	createUserHandler := command.NewCreateUserWithUoWHandler(uowFactory, eventBus)
	updateUserProfileHandler := command.NewUpdateUserProfileWithUoWHandler(uowFactory, eventBus)
	updateUserContactHandler := command.NewUpdateUserContactWithUoWHandler(uowFactory, eventBus)
	deleteUserHandler := command.NewDeleteUserWithUoWHandler(uowFactory, eventBus)

	// Initialize query handlers
	getUserHandler := query.NewGetUserHandler(userProjection)
	listUsersHandler := query.NewListUsersHandler(userProjection)
	searchUsersHandler := query.NewSearchUsersHandler(userProjection)

	// Initialize application service
	userService := services.NewUserService(
		createUserHandler,
		updateUserProfileHandler,
		updateUserContactHandler,
		deleteUserHandler,
		getUserHandler,
		listUsersHandler,
		searchUsersHandler,
	)

	// Start event bus
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		log.Fatal("Failed to start event bus:", err)
	}

	// Initialize HTTP controller
	userController := httpHandler.NewHTTPUserController(userService)

	// Setup HTTP routes
	mux := http.NewServeMux()

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			userController.CreateUser(w, r)
		case http.MethodGet:
			userController.ListUsers(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userController.GetUser(w, r)
		case http.MethodPut:
			userController.UpdateUser(w, r)
		case http.MethodDelete:
			userController.DeleteUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"whisko-petcare"}`))
	})

	// Start HTTP server
	go func() {
		port := getEnv("PORT", "8080")
		log.Printf("Server starting on port %s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	eventBus.Stop()
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
