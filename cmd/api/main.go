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
	"whisko-petcare/internal/infrastructure/cloudinary"
	httpHandler "whisko-petcare/internal/infrastructure/http"
	"whisko-petcare/internal/infrastructure/mongo"
	"whisko-petcare/internal/infrastructure/payos"
	"whisko-petcare/internal/infrastructure/projection"
	jwtutil "whisko-petcare/pkg/jwt"

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
	
	// Create concrete MongoDB user projection
	concreteUserProjection := projection.NewMongoUserProjection(database).(*projection.MongoUserProjection)
	userProjection := projection.UserProjection(concreteUserProjection)

	// Initialize JWT Manager
	jwtSecretKey := getEnv("JWT_SECRET_KEY", "your-super-secret-jwt-key-change-this-in-production-min-32-characters")
	tokenDuration, err := time.ParseDuration(getEnv("JWT_TOKEN_DURATION", "24h"))
	if err != nil {
		log.Printf("Invalid JWT_TOKEN_DURATION, using default 24h: %v", err)
		tokenDuration = 24 * time.Hour
	}
	jwtManager := jwtutil.NewJWTManager(jwtSecretKey, tokenDuration)
	log.Println("✅ JWT Manager initialized")

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

	// Initialize Cloudinary service
	var cloudinaryService *cloudinary.Service
	var cloudinaryHandler *cloudinary.Handler
	cloudinaryConfig, err := cloudinary.NewConfigFromEnv()
	if err != nil {
		log.Printf("⚠️  Warning: Cloudinary not configured: %v", err)
		log.Println("Image upload features will be disabled. To enable:")
		log.Println("  - Set CLOUDINARY_CLOUD_NAME in .env")
		log.Println("  - Set CLOUDINARY_API_KEY in .env")
		log.Println("  - Set CLOUDINARY_API_SECRET in .env")
	} else {
		log.Printf("📋 Cloudinary Config: CloudName=%s, APIKey=%s, Folder=%s", 
			cloudinaryConfig.CloudName, cloudinaryConfig.APIKey, cloudinaryConfig.UploadFolder)
		cloudinaryService, err = cloudinary.NewService(cloudinaryConfig)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to initialize Cloudinary service: %v", err)
		} else {
			cloudinaryHandler = cloudinary.NewHandler(cloudinaryService)
			log.Println("✅ Cloudinary service initialized")
		}
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

	eventBus.Subscribe("UserPasswordChanged", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserPasswordChanged(ctx, e.(*event.UserPasswordChanged))
		}))

	eventBus.Subscribe("UserRoleUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserRoleUpdated(ctx, e.(*event.UserRoleUpdated))
		}))

	eventBus.Subscribe("UserLoggedIn", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserLoggedIn(ctx, e.(*event.UserLoggedIn))
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

	eventBus.Subscribe("PetVaccinationAdded", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetVaccinationAdded(ctx, e.(*event.PetVaccinationAdded))
		}))

	eventBus.Subscribe("PetMedicalRecordAdded", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetMedicalRecordAdded(ctx, e.(*event.PetMedicalRecordAdded))
		}))

	eventBus.Subscribe("PetAllergyAdded", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetAllergyAdded(ctx, e.(*event.PetAllergyAdded))
		}))

	eventBus.Subscribe("PetAllergyRemoved", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return petProjection.HandlePetAllergyRemoved(ctx, e.(*event.PetAllergyRemoved))
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

	// Update image handlers
	updateUserImageHandler := command.NewUpdateUserImageWithUoWHandler(uowFactory, eventBus)

	// Initialize query handlers
	getUserHandler := query.NewGetUserHandler(userProjection)
	listUsersHandler := query.NewListUsersHandler(userProjection)
	searchUsersHandler := query.NewSearchUsersHandler(userProjection)

	// Initialize schedule command handlers FIRST (needed by payment confirmation)
	createScheduleHandler := command.NewCreateScheduleWithUoWHandler(uowFactory, eventBus)

	// Initialize payment command handlers with UoW
	createPaymentHandler := command.NewCreatePaymentWithUoWHandler(uowFactory, eventBus, payOSService)
	cancelPaymentHandler := command.NewCancelPaymentWithUoWHandler(uowFactory, eventBus, payOSService)
	confirmPaymentHandler := command.NewConfirmPaymentWithUoWHandler(uowFactory, eventBus, payOSService, createScheduleHandler)
	
	// Initialize payment query handlers
	getPaymentHandler := query.NewGetPaymentHandler(paymentProjection)
	getPaymentByOrderCodeHandler := query.NewGetPaymentByOrderCodeHandler(paymentProjection)
	listUserPaymentsHandler := query.NewListUserPaymentsHandler(paymentProjection)

	// Initialize pet command handlers
	createPetHandler := command.NewCreatePetWithUoWHandler(uowFactory, eventBus)
	updatePetHandler := command.NewUpdatePetWithUoWHandler(uowFactory, eventBus)
	deletePetHandler := command.NewDeletePetWithUoWHandler(uowFactory, eventBus)
	addPetVaccinationHandler := command.NewAddPetVaccinationWithUoWHandler(uowFactory, eventBus)
	addPetMedicalRecordHandler := command.NewAddPetMedicalRecordWithUoWHandler(uowFactory, eventBus)
	addPetAllergyHandler := command.NewAddPetAllergyWithUoWHandler(uowFactory, eventBus)
	removePetAllergyHandler := command.NewRemovePetAllergyWithUoWHandler(uowFactory, eventBus)

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

	// Continue with other schedule command handlers
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
		updateUserImageHandler,
		getUserHandler,
		listUsersHandler,
		searchUsersHandler,
	)

	petService := services.NewPetService(
		createPetHandler,
		updatePetHandler,
		deletePetHandler,
		addPetVaccinationHandler,
		addPetMedicalRecordHandler,
		addPetAllergyHandler,
		removePetAllergyHandler,
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

	// Initialize auth command handlers
	registerHandler := command.NewRegisterUserWithUoWHandler(uowFactory, eventBus)
	changePasswordHandler := command.NewChangeUserPasswordWithUoWHandler(uowFactory, eventBus)
	recordLoginHandler := command.NewRecordUserLoginWithUoWHandler(uowFactory, eventBus)

	// Initialize HTTP controllers
	userController := httpHandler.NewHTTPUserController(userService, cloudinaryService)
	authController := httpHandler.NewHTTPAuthController(registerHandler, changePasswordHandler, recordLoginHandler, concreteUserProjection, jwtManager)
	paymentController := httpHandler.NewHTTPPaymentController(
		createPaymentHandler,
		cancelPaymentHandler,
		confirmPaymentHandler,
		getPaymentHandler,
		getPaymentByOrderCodeHandler,
		listUserPaymentsHandler,
		payOSService,
	)
	petController := httpHandler.NewHTTPPetController(petService, cloudinaryService)
	vendorController := httpHandler.NewVendorController(vendorService, cloudinaryService)
	serviceController := httpHandler.NewHTTPServiceController(serviceService, cloudinaryService)
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
		// Check for nested routes under /users/{userID}/...
		if strings.Contains(r.URL.Path, "/pets") && r.Method == http.MethodGet {
			petController.ListUserPets(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/schedules") && r.Method == http.MethodGet {
			scheduleController.ListUserSchedules(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "/vendor-staffs") && r.Method == http.MethodGet {
			vendorStaffController.ListVendorStaffByUser(w, r)
			return
		}
		// Update image endpoint: PUT /users/{id}/image
		if strings.HasSuffix(r.URL.Path, "/image") && r.Method == http.MethodPut {
			userController.UpdateUserImage(w, r)
			return
		}
		
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

	// Payment routes - Order matters! More specific routes first
	
	// PayOS webhook and return URLs (most specific first)
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

	// Payment base route (POST only - exact match)
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments" {
			// If path is not exactly "/payments", let it fall through to /payments/ handler
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodPost:
			paymentController.CreatePayment(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Payment query routes (specific paths) - These must come AFTER /payments base route
	// Go's ServeMux will match the longest pattern first
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

	// Payment by ID routes (with trailing slash - catches /payments/{id})
	// This will match anything like /payments/xxx that wasn't caught by more specific routes
	mux.HandleFunc("/payments/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch r.Method {
		case http.MethodGet:
			paymentController.GetPayment(w, r)
		case http.MethodPut:
			if strings.HasSuffix(path, "/cancel") {
				paymentController.CancelPayment(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Pet routes
	mux.HandleFunc("/pets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			petController.CreatePet(w, r)
		case http.MethodGet:
			petController.ListPets(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/pets/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/pets/")
		parts := strings.Split(path, "/")
		
		// Check for health endpoints: /pets/{id}/vaccinations, /pets/{id}/medical-records, /pets/{id}/allergies
		if len(parts) >= 2 {
			resource := parts[1]
			
			switch resource {
			case "vaccinations":
				if r.Method == http.MethodPost {
					petController.AddPetVaccination(w, r)
					return
				}
			case "medical-records":
				if r.Method == http.MethodPost {
					petController.AddPetMedicalRecord(w, r)
					return
				}
			case "allergies":
				if r.Method == http.MethodPost {
					petController.AddPetAllergy(w, r)
					return
				}
				// Handle DELETE /pets/{id}/allergies/{allergy_id}
				if r.Method == http.MethodDelete && len(parts) >= 3 {
					petController.RemovePetAllergy(w, r)
					return
				}
			}
		}
		
		// Default pet routes
		switch r.Method {
		case http.MethodGet:
			petController.GetPet(w, r)
		case http.MethodPut:
			petController.UpdatePet(w, r)
		case http.MethodDelete:
			petController.DeletePet(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Vendor routes
	mux.HandleFunc("/vendors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			vendorController.CreateVendor(w, r)
		case http.MethodGet:
			vendorController.ListVendors(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/vendors/", func(w http.ResponseWriter, r *http.Request) {
		// Check for /vendors/{vendorID}/services
		if strings.Contains(r.URL.Path, "/services") && r.Method == http.MethodGet {
			serviceController.ListVendorServices(w, r)
			return
		}
		// Check for /vendors/{vendorID}/staff
		if strings.Contains(r.URL.Path, "/staff") && r.Method == http.MethodGet {
			vendorStaffController.ListVendorStaffByVendor(w, r)
			return
		}
		// Check for /vendors/{shopID}/schedules
		if strings.Contains(r.URL.Path, "/schedules") && r.Method == http.MethodGet {
			scheduleController.ListShopSchedules(w, r)
			return
		}
		
		switch r.Method {
		case http.MethodGet:
			vendorController.GetVendor(w, r)
		case http.MethodPut:
			vendorController.UpdateVendor(w, r)
		case http.MethodDelete:
			vendorController.DeleteVendor(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Service routes
	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			serviceController.CreateService(w, r)
		case http.MethodGet:
			serviceController.ListServices(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/services/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			serviceController.GetService(w, r)
		case http.MethodPut:
			serviceController.UpdateService(w, r)
		case http.MethodDelete:
			serviceController.DeleteService(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Schedule routes
	mux.HandleFunc("/schedules", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			scheduleController.CreateSchedule(w, r)
		case http.MethodGet:
			scheduleController.ListSchedules(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/schedules/", func(w http.ResponseWriter, r *http.Request) {
		// Check for /schedules/{id}/status
		if strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodPut {
			scheduleController.ChangeScheduleStatus(w, r)
			return
		}
		// Check for /schedules/{id}/complete
		if strings.HasSuffix(r.URL.Path, "/complete") && r.Method == http.MethodPost {
			scheduleController.CompleteSchedule(w, r)
			return
		}
		// Check for /schedules/{id}/cancel
		if strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == http.MethodPost {
			scheduleController.CancelSchedule(w, r)
			return
		}
		
		switch r.Method {
		case http.MethodGet:
			scheduleController.GetSchedule(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Vendor Staff routes
	mux.HandleFunc("/vendor-staffs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			vendorStaffController.CreateVendorStaff(w, r)
		case http.MethodGet:
			vendorStaffController.ListVendorStaffs(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/vendor-staffs/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vendorStaffController.GetVendorStaff(w, r)
		case http.MethodDelete:
			vendorStaffController.DeleteVendorStaff(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Auth routes
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authController.Register(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authController.Login(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/auth/change-password", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authController.ChangePassword(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Cloudinary image routes
	if cloudinaryHandler != nil {
		mux.HandleFunc("/api/images/upload", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				cloudinaryHandler.HandleUploadImage(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

		mux.HandleFunc("/api/images/delete", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				cloudinaryHandler.HandleDeleteImage(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

		mux.HandleFunc("/api/images/transform", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				cloudinaryHandler.HandleGetTransformedURL(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

		log.Println("✅ Cloudinary routes registered:")
		log.Println("   POST   /api/images/upload")
		log.Println("   DELETE /api/images/delete")
		log.Println("   POST   /api/images/transform")
	}

	// API routes for services
	mux.HandleFunc("/api/services/vendor/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Extract vendor ID from path: /api/services/vendor/{vendorID}
			serviceController.ListVendorServices(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	log.Println("✅ Service API routes registered:")
	log.Println("   GET    /api/services/vendor/{vendorID}")

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"whisko-petcare"}`))
	})

	// Start payment expiry background service
	paymentExpiryService := services.NewPaymentExpiryService(uowFactory, eventBus, payOSService)
	go paymentExpiryService.Start(context.Background())

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
	paymentExpiryService.Stop()
	eventBus.Stop()
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
