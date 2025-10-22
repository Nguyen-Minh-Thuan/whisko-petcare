package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/infrastructure/bus"
	httpHandler "whisko-petcare/internal/infrastructure/http"
	"whisko-petcare/internal/infrastructure/mongo"
	"whisko-petcare/internal/infrastructure/payos"
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

	// Initialize PayOS service
	payOSConfig := &payos.Config{
		ClientID:    getEnv("PAYOS_CLIENT_ID", ""),
		APIKey:      getEnv("PAYOS_API_KEY", ""),
		ChecksumKey: getEnv("PAYOS_CHECKSUM_KEY", ""),
		PartnerCode: getEnv("PAYOS_PARTNER_CODE", ""),
		ReturnURL:   getEnv("PAYOS_RETURN_URL", "http://localhost:8080/payments/return"),
		CancelURL:   getEnv("PAYOS_CANCEL_URL", "http://localhost:8080/payments/cancel"),
	}
	payOSService, err := payos.NewService(payOSConfig)
	if err != nil {
		log.Fatal("Failed to initialize PayOS service:", err)
	}

	// Initialize projections
	paymentProjection := projection.NewMongoPaymentProjection(database)
	petProjection := projection.NewMongoPetProjection(database)
	vendorProjection := projection.NewMongoVendorProjection(database)
	serviceProjection := projection.NewMongoServiceProjection(database)
	scheduleProjection := projection.NewMongoScheduleProjection(database)
	vendorStaffProjection := projection.NewMongoVendorStaffProjection(database)

	// Initialize Unit of Work factory
	uowFactory := mongo.NewMongoUnitOfWorkFactory(mongoClient.GetClient(), database)

	// Subscribe user projection to events
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

	// Subscribe payment projection to events
	eventBus.Subscribe("PaymentCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return paymentProjection.HandlePaymentCreated(ctx, e.(*event.PaymentCreated))
		}))

	eventBus.Subscribe("PaymentUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return paymentProjection.HandlePaymentUpdated(ctx, e.(*event.PaymentUpdated))
		}))

	eventBus.Subscribe("PaymentStatusChanged", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return paymentProjection.HandlePaymentStatusChanged(ctx, e.(*event.PaymentStatusChanged))
		}))

	// Subscribe pet projection to events
	eventBus.Subscribe("PetCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetCreated(ctx, e.(*event.PetCreated))
		}))

	eventBus.Subscribe("PetUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetUpdated(ctx, e.(*event.PetUpdated))
		}))

	eventBus.Subscribe("PetDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetDeleted(ctx, e.(*event.PetDeleted))
		}))

	// Subscribe vendor projection to events
	eventBus.Subscribe("VendorCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return vendorProjection.HandleVendorCreated(ctx, *e.(*event.VendorCreated))
		}))

	eventBus.Subscribe("VendorUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return vendorProjection.HandleVendorUpdated(ctx, *e.(*event.VendorUpdated))
		}))

	eventBus.Subscribe("VendorDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return vendorProjection.HandleVendorDeleted(ctx, *e.(*event.VendorDeleted))
		}))

	// Subscribe service projection to events
	eventBus.Subscribe("ServiceCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return serviceProjection.HandleServiceCreated(ctx, *e.(*event.ServiceCreated))
		}))

	eventBus.Subscribe("ServiceUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return serviceProjection.HandleServiceUpdated(ctx, *e.(*event.ServiceUpdated))
		}))

	eventBus.Subscribe("ServiceDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return serviceProjection.HandleServiceDeleted(ctx, *e.(*event.ServiceDeleted))
		}))

	// Subscribe schedule projection to events
	eventBus.Subscribe("ScheduleCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return scheduleProjection.HandleScheduleCreated(ctx, *e.(*event.ScheduleCreated))
		}))

	eventBus.Subscribe("ScheduleStatusChanged", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return scheduleProjection.HandleScheduleStatusChanged(ctx, *e.(*event.ScheduleStatusChanged))
		}))

	eventBus.Subscribe("ScheduleCompleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return scheduleProjection.HandleScheduleCompleted(ctx, *e.(*event.ScheduleCompleted))
		}))

	eventBus.Subscribe("ScheduleCancelled", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return scheduleProjection.HandleScheduleCancelled(ctx, *e.(*event.ScheduleCancelled))
		}))

	// Subscribe vendor staff projection to events
	eventBus.Subscribe("VendorStaffCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return vendorStaffProjection.HandleVendorStaffCreated(ctx, *e.(*event.VendorStaffCreated))
		}))

	eventBus.Subscribe("VendorStaffDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return vendorStaffProjection.HandleVendorStaffDeleted(ctx, *e.(*event.VendorStaffDeleted))
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

	// Initialize payment command handlers with UoW
	createPaymentHandler := command.NewCreatePaymentWithUoWHandler(uowFactory, eventBus, payOSService)
	cancelPaymentHandler := command.NewCancelPaymentWithUoWHandler(uowFactory, eventBus, payOSService)
	confirmPaymentHandler := command.NewConfirmPaymentWithUoWHandler(uowFactory, eventBus, payOSService)
	
	// Initialize payment query handlers
	getPaymentHandler := query.NewGetPaymentHandler(paymentProjection)
	getPaymentByOrderCodeHandler := query.NewGetPaymentByOrderCodeHandler(paymentProjection)
	listUserPaymentsHandler := query.NewListUserPaymentsHandler(paymentProjection)

	// Initialize pet command handlers
	createPetHandler := command.NewCreatePetWithUoWHandler(uowFactory, eventBus)
	updatePetHandler := command.NewUpdatePetWithUoWHandler(uowFactory, eventBus)
	deletePetHandler := command.NewDeletePetWithUoWHandler(uowFactory, eventBus)

	// Initialize pet query handlers
	getPetHandler := query.NewGetPetHandler(petProjection)
	listUserPetsHandler := query.NewListUserPetsHandler(petProjection)
	listPetsHandler := query.NewListPetsHandler(petProjection)

	// Initialize vendor command handlers
	createVendorHandler := command.NewCreateVendorWithUoWHandler(uowFactory, eventBus)
	updateVendorHandler := command.NewUpdateVendorWithUoWHandler(uowFactory, eventBus)
	deleteVendorHandler := command.NewDeleteVendorWithUoWHandler(uowFactory, eventBus)

	// Initialize vendor query handlers
	getVendorHandler := query.NewGetVendorHandler(vendorProjection)
	listVendorsHandler := query.NewListVendorsHandler(vendorProjection)

	// Initialize service command handlers
	createServiceHandler := command.NewCreateServiceWithUoWHandler(uowFactory, eventBus)
	updateServiceHandler := command.NewUpdateServiceWithUoWHandler(uowFactory, eventBus)
	deleteServiceHandler := command.NewDeleteServiceWithUoWHandler(uowFactory, eventBus)

	// Initialize service query handlers
	getServiceHandler := query.NewGetServiceHandler(serviceProjection)
	listVendorServicesHandler := query.NewListVendorServicesHandler(serviceProjection)
	listServicesHandler := query.NewListServicesHandler(serviceProjection)

	// Initialize schedule command handlers
	createScheduleHandler := command.NewCreateScheduleWithUoWHandler(uowFactory, eventBus)
	changeScheduleStatusHandler := command.NewChangeScheduleStatusWithUoWHandler(uowFactory, eventBus)
	completeScheduleHandler := command.NewCompleteScheduleWithUoWHandler(uowFactory, eventBus)
	cancelScheduleHandler := command.NewCancelScheduleWithUoWHandler(uowFactory, eventBus)

	// Initialize schedule query handlers
	getScheduleHandler := query.NewGetScheduleHandler(scheduleProjection)
	listUserSchedulesHandler := query.NewListUserSchedulesHandler(scheduleProjection)
	listShopSchedulesHandler := query.NewListShopSchedulesHandler(scheduleProjection)
	listSchedulesHandler := query.NewListSchedulesHandler(scheduleProjection)

	// Initialize vendor staff command handlers
	createVendorStaffHandler := command.NewCreateVendorStaffWithUoWHandler(uowFactory, eventBus)
	deleteVendorStaffHandler := command.NewDeleteVendorStaffWithUoWHandler(uowFactory, eventBus)

	// Initialize vendor staff query handlers
	getVendorStaffHandler := query.NewGetVendorStaffHandler(vendorStaffProjection)
	listVendorStaffByVendorHandler := query.NewListVendorStaffByVendorHandler(vendorStaffProjection)
	listVendorStaffByUserHandler := query.NewListVendorStaffByUserHandler(vendorStaffProjection)
	listVendorStaffsHandler := query.NewListVendorStaffsHandler(vendorStaffProjection)

	// Initialize application services
	userService := services.NewUserService(
		createUserHandler,
		updateUserProfileHandler,
		updateUserContactHandler,
		deleteUserHandler,
		getUserHandler,
		listUsersHandler,
		searchUsersHandler,
	)

	petService := services.NewPetService(
		createPetHandler,
		updatePetHandler,
		deletePetHandler,
		getPetHandler,
		listUserPetsHandler,
		listPetsHandler,
	)

	vendorService := services.NewVendorService(
		createVendorHandler,
		updateVendorHandler,
		deleteVendorHandler,
		getVendorHandler,
		listVendorsHandler,
	)

	serviceService := services.NewServiceService(
		createServiceHandler,
		updateServiceHandler,
		deleteServiceHandler,
		getServiceHandler,
		listVendorServicesHandler,
		listServicesHandler,
	)

	scheduleService := services.NewScheduleService(
		createScheduleHandler,
		changeScheduleStatusHandler,
		completeScheduleHandler,
		cancelScheduleHandler,
		getScheduleHandler,
		listUserSchedulesHandler,
		listShopSchedulesHandler,
		listSchedulesHandler,
	)

	vendorStaffService := services.NewVendorStaffService(
		createVendorStaffHandler,
		deleteVendorStaffHandler,
		getVendorStaffHandler,
		listVendorStaffByVendorHandler,
		listVendorStaffByUserHandler,
		listVendorStaffsHandler,
	)

	// Start event bus
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventBus.Start(ctx); err != nil {
		log.Fatal("Failed to start event bus:", err)
	}

	// Initialize HTTP controllers
	userController := httpHandler.NewHTTPUserController(userService)
	paymentController := httpHandler.NewHTTPPaymentController(
		createPaymentHandler,
		cancelPaymentHandler,
		confirmPaymentHandler,
		getPaymentHandler,
		getPaymentByOrderCodeHandler,
		listUserPaymentsHandler,
		payOSService,
	)
	petController := httpHandler.NewHTTPPetController(petService)
	vendorController := httpHandler.NewVendorController(vendorService)
	serviceController := httpHandler.NewHTTPServiceController(serviceService)
	scheduleController := httpHandler.NewScheduleController(scheduleService)
	vendorStaffController := httpHandler.NewVendorStaffController(vendorStaffService)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// User routes
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

	// Payment routes
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			paymentController.CreatePayment(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/payments/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.Contains(r.URL.Path, "/cancel") {
				// This is a cancel request
				paymentController.CancelPayment(w, r)
			} else {
				// This is a get request
				paymentController.GetPayment(w, r)
			}
		case http.MethodPut:
			if strings.Contains(r.URL.Path, "/cancel") {
				paymentController.CancelPayment(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Payment special routes
	mux.HandleFunc("/payments/order/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paymentController.GetPaymentByOrderCode(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/payments/user/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paymentController.ListUserPayments(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// PayOS webhook and return URLs
	mux.HandleFunc("/payments/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			paymentController.WebhookHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/payments/return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paymentController.ReturnHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/payments/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paymentController.CancelHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Pet routes
	mux.HandleFunc("POST /pets", petController.CreatePet)
	mux.HandleFunc("GET /pets/{id}", petController.GetPet)
	mux.HandleFunc("GET /pets", petController.ListPets)
	mux.HandleFunc("GET /users/{userID}/pets", petController.ListUserPets)
	mux.HandleFunc("PUT /pets/{id}", petController.UpdatePet)
	mux.HandleFunc("DELETE /pets/{id}", petController.DeletePet)

	// Vendor routes
	mux.HandleFunc("POST /vendors", vendorController.CreateVendor)
	mux.HandleFunc("GET /vendors/{id}", vendorController.GetVendor)
	mux.HandleFunc("GET /vendors", vendorController.ListVendors)
	mux.HandleFunc("PUT /vendors/{id}", vendorController.UpdateVendor)
	mux.HandleFunc("DELETE /vendors/{id}", vendorController.DeleteVendor)

	// Service routes
	mux.HandleFunc("POST /services", serviceController.CreateService)
	mux.HandleFunc("GET /services/{id}", serviceController.GetService)
	mux.HandleFunc("GET /services", serviceController.ListServices)
	mux.HandleFunc("GET /vendors/{vendorID}/services", serviceController.ListVendorServices)
	mux.HandleFunc("PUT /services/{id}", serviceController.UpdateService)
	mux.HandleFunc("DELETE /services/{id}", serviceController.DeleteService)

	// Schedule routes
	mux.HandleFunc("POST /schedules", scheduleController.CreateSchedule)
	mux.HandleFunc("GET /schedules/{id}", scheduleController.GetSchedule)
	mux.HandleFunc("GET /schedules", scheduleController.ListSchedules)
	mux.HandleFunc("GET /users/{userID}/schedules", scheduleController.ListUserSchedules)
	mux.HandleFunc("GET /vendors/{shopID}/schedules", scheduleController.ListShopSchedules)
	mux.HandleFunc("PUT /schedules/{id}/status", scheduleController.ChangeScheduleStatus)
	mux.HandleFunc("POST /schedules/{id}/complete", scheduleController.CompleteSchedule)
	mux.HandleFunc("POST /schedules/{id}/cancel", scheduleController.CancelSchedule)

	// Vendor Staff routes
	mux.HandleFunc("POST /vendor-staffs", vendorStaffController.CreateVendorStaff)
	mux.HandleFunc("GET /vendor-staffs/{userID}/{vendorID}", vendorStaffController.GetVendorStaff)
	mux.HandleFunc("GET /vendor-staffs", vendorStaffController.ListVendorStaffs)
	mux.HandleFunc("GET /vendors/{vendorID}/staff", vendorStaffController.ListVendorStaffByVendor)
	mux.HandleFunc("GET /users/{userID}/vendor-staffs", vendorStaffController.ListVendorStaffByUser)
	mux.HandleFunc("DELETE /vendor-staffs/{userID}/{vendorID}", vendorStaffController.DeleteVendorStaff)

	// Health check
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
