package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"api-core/config"
	"api-core/internal/routes"
	"api-core/internal/schedules"
	"api-core/internal/wire"
	"api-core/pkg/actionEvent"
	"api-core/pkg/cache"
	"api-core/pkg/cron"
	"api-core/pkg/exception"
	"api-core/pkg/i18n"
	"api-core/pkg/logger"
	middlewarePkg "api-core/pkg/middleware"
	socketPkg "api-core/pkg/socket"
	"api-core/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	loadEnvironment()

	// Initialize logger
	initLogger()

	logger.Info("Starting ApiCore application...")

	// Initialize i18n
	initI18n()

	// Initialize validation messages
	initValidation()

	// Initialize Loki events
	initActionEvents()

	// Connect to database
	db := initDatabase()

	// Connect to cache
	cacheClient := initCache()

	// Initialize dependencies
	controllers := initDependencies(db, cacheClient)

	// Initialize schedule manager
	scheduleManager := initScheduleManager()

	// Initialize socket hub
	socketHub := initSocketHub()

	// Setup router and routes
	r := setupRouter(controllers, socketHub)

	// Start schedule manager
	startScheduleManager(scheduleManager)

	// Start server
	startServer(r)
}

// loadEnvironment loads environment variables from .env file
func loadEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

// initLogger initializes the logger
func initLogger() {
	// Load logger config từ environment variables
	loggerConfig := config.LoadLoggerConfig()

	// Validate config
	if err := loggerConfig.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid logger config: %v", err))
	}

	// Convert to logger.Config và initialize
	if err := logger.Init(loggerConfig.ToLoggerConfig()); err != nil {
		panic(err)
	}
}

// initI18n initializes internationalization
func initI18n() {
	if err := i18n.Init(i18n.Config{
		TranslationsDir: "translations",
		Languages:       []string{"en", "vi"},
		FallbackLang:    "en",
	}); err != nil {
		logger.Warnf("Failed to initialize i18n: %v (using default messages)", err)
	} else {
		logger.Info("I18n initialized successfully")
	}
}

// initValidation initializes validation messages
func initValidation() {
	validator.InitValidationMessages(i18n.GetTranslator())
	logger.Info("Validation messages initialized successfully")
}

// initActionEvents initializes action events
func initActionEvents() {
	actionEventConfig := config.LoadActionEventConfig()
	if !actionEventConfig.Enabled {
		logger.Info("Action events disabled")
		return
	}

	// Create Loki client
	lokiClient := actionEvent.NewLokiClient(actionEventConfig.LokiURL, map[string]string{
		"environment": actionEventConfig.Environment,
		"host":        "apicore",
	})

	// Initialize action event service
	actionEvent.Init(lokiClient)
	logger.Info("Action events initialized successfully")
}

// initDatabase connects to the database
func initDatabase() *gorm.DB {
	dbConfig := config.GetDefaultDatabaseConfig()
	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Database connected successfully")
	return db
}

// initCache connects to the cache
func initCache() cache.Cache {
	cacheConfig := config.GetDefaultCacheConfig()
	cacheClient, err := config.ConnectCache(cacheConfig)
	if err != nil {
		logger.Warnf("Failed to connect to cache: %v (using no-op cache)", err)
		// Use no-op cache - app vẫn chạy nhưng không cache
		cacheClient = cache.NewNoopCache()
	} else {
		logger.Info("Cache connected successfully")
	}
	return cacheClient
}

// initDependencies initializes all dependencies using wire
func initDependencies(db *gorm.DB, cacheClient cache.Cache) *routes.Controllers {
	controllers, err := wire.InitializeApp(db, cacheClient)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	logger.Info("Dependencies initialized successfully")
	return controllers
}

// initScheduleManager initializes the schedule manager
func initScheduleManager() *schedules.ScheduleManager {
	// Create Redis client for schedule manager
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warnf("Failed to connect to Redis for schedule manager: %v", err)
		logger.Info("Using memory lock manager for schedule manager")

		// Close Redis connection if not available
		rdb.Close()

		// Use memory lock manager if Redis is not available
		lockManager := cron.NewMemoryLockManager()
		manager, err := schedules.InitScheduleManager(lockManager)
		if err != nil {
			logger.Warnf("Failed to initialize schedule manager: %v", err)
			return nil
		}
		return manager
	}

	// Use Redis lock manager for multi-container deployment
	lockManager := cron.NewRedisLockManager(rdb, "api-core:cron:")
	manager, err := schedules.InitScheduleManager(lockManager)
	if err != nil {
		logger.Warnf("Failed to initialize schedule manager: %v", err)
		rdb.Close()
		return nil
	}

	logger.Info("Schedule manager initialized successfully")
	return manager
}

// initSocketHub initializes the WebSocket hub
func initSocketHub() *socketPkg.Hub {
	hub := socketPkg.NewHub()

	// Start the hub in a goroutine
	go hub.Run()

	logger.Info("WebSocket hub initialized successfully")
	return hub
}

// setupRouter sets up the router and all routes
func setupRouter(controllers *routes.Controllers, socketHub *socketPkg.Hub) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID) // Tạo unique ID cho mỗi request
	r.Use(logger.Middleware())  // Log requests/responses với đầy đủ thông tin
	r.Use(i18n.Middleware)      // Tự động detect và set language vào context

	// Custom headers middleware
	r.Use(middlewarePkg.CORSHeaders())     // CORS headers
	r.Use(middlewarePkg.SecurityHeaders()) // Security headers

	// Custom headers for specific endpoints
	r.Use(middlewarePkg.CustomHeaders(map[string]string{
		// Headers will be set from environment variables
	}))
	r.Use(exception.RecoveryMiddleware) // Recover từ panic với custom exception handling

	// Setup documentation routes
	setupDocumentationRoutes(r)

	// Setup static file routes
	setupStaticFileRoutes(r)

	// Register all API routes
	routes.RegisterRoutes(r, controllers)

	// Register WebSocket routes
	socketPkg.RegisterRoutes(r, socketHub)

	return r
}

// setupDocumentationRoutes sets up documentation routes
func setupDocumentationRoutes(r *chi.Mux) {
	workDir, _ := os.Getwd()
	docsDir := http.Dir(filepath.Join(workDir, "docs"))

	// Redirect root to docs
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})

	// Docs home page
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "docs", "index.html"))
	})

	// Swagger UI
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "docs", "swagger.html"))
	})

	// Swagger JSON
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(workDir, "docs", "swagger.json"))
	})

	// Static files in docs
	r.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/docs/", http.FileServer(docsDir)).ServeHTTP(w, r)
	})
}

// setupStaticFileRoutes sets up static file routes
func setupStaticFileRoutes(r *chi.Mux) {
	workDir, _ := os.Getwd()

	// Static files for storages (avatars, etc.)
	storageDir := http.Dir(filepath.Join(workDir, "storages/app"))
	r.Get("/storages/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/storages/", http.FileServer(storageDir)).ServeHTTP(w, r)
	})

	// WebSocket test page
	r.Get("/test-socket", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "examples", "test_socket.html"))
	})
}

// startScheduleManager starts the schedule manager
func startScheduleManager(manager *schedules.ScheduleManager) {
	if manager == nil {
		logger.Warn("Schedule manager not initialized, skipping...")
		return
	}

	// Create a background context that won't be cancelled
	ctx := context.Background()
	if err := manager.Start(ctx); err != nil {
		logger.Warnf("Failed to start schedule manager: %v", err)
		return
	}

	logger.Info("Schedule manager started successfully")
}

// startServer starts the HTTP server
func startServer(r *chi.Mux) {
	logger.Info("Server starting on :3000")
	logger.Info("Documentation: http://localhost:3000/docs")
	logger.Info("Swagger UI: http://localhost:3000/swagger")
	logger.Info("WebSocket Test: http://localhost:3000/test-socket")
	logger.Info("WebSocket Endpoint: ws://localhost:3000/ws")

	if err := http.ListenAndServe(":3000", r); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
