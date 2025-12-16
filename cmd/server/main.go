package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"cold-backend/internal/auth"
	"cold-backend/internal/config"
	"cold-backend/internal/database"
	"cold-backend/internal/db"
	h "cold-backend/internal/http"
	"cold-backend/internal/handlers"
	"cold-backend/internal/health"
	"cold-backend/internal/middleware"
	"cold-backend/internal/monitoring"
	"cold-backend/internal/repositories"
	"cold-backend/internal/services"
	"cold-backend/internal/sms"
)

func main() {
	// Parse command-line flags
	mode := flag.String("mode", "employee", "Server mode: employee or customer")
	port := flag.Int("port", 0, "Server port (overrides config)")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Override port if specified
	if *port != 0 {
		cfg.Server.Port = *port
	} else {
		// Set default ports based on mode
		if *mode == "customer" {
			cfg.Server.Port = 8081
		}
		// Employee mode uses config.yaml port (8080)
	}

	// Connect to database
	pool := db.Connect(cfg)
	defer pool.Close()

	// Run database migrations
	// This automatically creates all required tables on startup
	log.Println("Running database migrations...")
	migrator := database.NewMigrator(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrator.RunMigrations(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize health checker
	healthChecker := health.NewHealthChecker(pool)

	// Start monitoring dashboard server in background
	go monitoring.NewMonitoringServer(pool, 9090).Start()

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(pool)
	customerRepo := repositories.NewCustomerRepository(pool)
	entryRepo := repositories.NewEntryRepository(pool)
	entryEventRepo := repositories.NewEntryEventRepository(pool)
	roomEntryRepo := repositories.NewRoomEntryRepository(pool)
	systemSettingRepo := repositories.NewSystemSettingRepository(pool)
	rentPaymentRepo := repositories.NewRentPaymentRepository(pool)
	gatePassRepo := repositories.NewGatePassRepository(pool)
	invoiceRepo := repositories.NewInvoiceRepository(pool)
	loginLogRepo := repositories.NewLoginLogRepository(pool)
	roomEntryEditLogRepo := repositories.NewRoomEntryEditLogRepository(pool)
	adminActionLogRepo := repositories.NewAdminActionLogRepository(pool)
	gatePassPickupRepo := repositories.NewGatePassPickupRepository(pool)

	// Initialize middleware (needed for both modes)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, userRepo)
	corsMiddleware := middleware.NewCORS(cfg)
	pageHandler := handlers.NewPageHandler()
	healthHandler := handlers.NewHealthHandler(healthChecker)

	var handler http.Handler

	if *mode == "customer" {
		log.Println("Starting in CUSTOMER PORTAL mode")

		// Initialize OTP repository and SMS service
		otpRepo := repositories.NewOTPRepository(pool)

		// Use MockSMSService for testing (prints OTP to console)
		// For production, use: smsService := sms.NewFast2SMSService(cfg.SMS.APIKey)
		smsService := sms.NewMockSMSService()

		// Initialize OTP service
		otpService := services.NewOTPService(otpRepo, customerRepo, smsService)

		// Initialize customer portal service
		customerPortalService := services.NewCustomerPortalService(
			customerRepo,
			entryRepo,
			roomEntryRepo,
			gatePassRepo,
			rentPaymentRepo,
		)

		// Initialize customer portal handler
		customerPortalHandler := handlers.NewCustomerPortalHandler(
			otpService,
			customerPortalService,
			jwtManager,
		)

		// Create customer router
		router := h.NewCustomerRouter(customerPortalHandler, pageHandler, healthHandler, authMiddleware)

		// Wrap with panic recovery and metrics middleware
		handler = middleware.PanicRecovery(middleware.MetricsMiddleware(corsMiddleware(router)))

	} else {
		log.Println("Starting in EMPLOYEE mode")

		// Initialize services (employee mode)
		userService := services.NewUserService(userRepo, jwtManager)
		customerService := services.NewCustomerService(customerRepo)
		entryService := services.NewEntryService(entryRepo, customerRepo, entryEventRepo)
		roomEntryService := services.NewRoomEntryService(roomEntryRepo, entryRepo, entryEventRepo)
		systemSettingService := services.NewSystemSettingService(systemSettingRepo)
		rentPaymentService := services.NewRentPaymentService(rentPaymentRepo)
		invoiceService := services.NewInvoiceService(invoiceRepo)
		gatePassService := services.NewGatePassService(gatePassRepo, entryRepo, entryEventRepo, gatePassPickupRepo, roomEntryRepo)

		// Initialize handlers (employee mode)
		userHandler := handlers.NewUserHandler(userService, adminActionLogRepo)
		authHandler := handlers.NewAuthHandler(userService, loginLogRepo)
		customerHandler := handlers.NewCustomerHandler(customerService)
		entryHandler := handlers.NewEntryHandler(entryService)
		roomEntryHandler := handlers.NewRoomEntryHandler(roomEntryService, roomEntryEditLogRepo)
		entryEventHandler := handlers.NewEntryEventHandler(entryEventRepo)
		systemSettingHandler := handlers.NewSystemSettingHandler(systemSettingService)
		rentPaymentHandler := handlers.NewRentPaymentHandler(rentPaymentService)
		invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
		loginLogHandler := handlers.NewLoginLogHandler(loginLogRepo)
		roomEntryEditLogHandler := handlers.NewRoomEntryEditLogHandler(roomEntryEditLogRepo)
		adminActionLogHandler := handlers.NewAdminActionLogHandler(adminActionLogRepo)
		gatePassHandler := handlers.NewGatePassHandler(gatePassService, adminActionLogRepo)

		// Create employee router
		router := h.NewRouter(userHandler, authHandler, customerHandler, entryHandler, roomEntryHandler, entryEventHandler, systemSettingHandler, rentPaymentHandler, invoiceHandler, loginLogHandler, roomEntryEditLogHandler, adminActionLogHandler, gatePassHandler, pageHandler, healthHandler, authMiddleware)

		// Wrap with panic recovery and metrics middleware
		handler = middleware.PanicRecovery(middleware.MetricsMiddleware(corsMiddleware(router)))
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server running on %s (mode: %s)", addr, *mode)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
