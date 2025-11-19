package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"api-core/config"
	"api-core/internal/routes"
	"api-core/internal/schedules"
	"api-core/internal/wire"
	"api-core/pkg/actionEvent"
	"api-core/pkg/cache"
	"api-core/pkg/cron"
	"api-core/pkg/exception"
	"api-core/pkg/fcm"
	"api-core/pkg/i18n"
	"api-core/pkg/logger"
	middlewarePkg "api-core/pkg/middleware"
	socketPkg "api-core/pkg/socket"
	"api-core/pkg/utils"
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

	// Initialize FCM client (only for test pages in development)
	fcmClient := initFCM()

	// Setup router and routes
	r := setupRouter(controllers, socketHub, fcmClient)

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

// initFCM initializes FCM client (optional, for test pages)
func initFCM() *fcm.Client {
	// Only initialize in development
	appEnv := utils.GetEnv("APP_ENV", "production")
	if appEnv != "development" {
		logger.Info("FCM client initialization skipped (not in development mode)")
		return nil
	}

	credentialsFile := utils.GetEnv("FIREBASE_CREDENTIALS_FILE", "keys/firebase-credentials.json")
	timeoutSeconds := utils.GetEnvInt("FCM_TIMEOUT", 10)

	// Check if credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		logger.Warnf("FCM credentials file not found: %s. FCM test APIs will not work.", credentialsFile)
		return nil
	}

	config := &fcm.Config{
		CredentialsFile: credentialsFile,
		Timeout:         time.Duration(timeoutSeconds) * time.Second,
	}

	client, err := fcm.NewClient(config)
	if err != nil {
		logger.Warnf("Failed to initialize FCM client: %v. FCM test APIs will not work.", err)
		return nil
	}

	logger.Info("FCM client initialized successfully")
	return client
}

// setupRouter sets up the router and all routes
func setupRouter(controllers *routes.Controllers, socketHub *socketPkg.Hub, fcmClient *fcm.Client) *chi.Mux {
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

	// Setup test pages (only in development)
	initTestPages(r, fcmClient)

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
}

// initTestPages sets up test pages (only available in development environment)
func initTestPages(r *chi.Mux, fcmClient *fcm.Client) {
	// Only initialize test pages in development environment
	appEnv := utils.GetEnv("APP_ENV", "production")
	if appEnv != "development" {
		logger.Info("Test pages disabled (APP_ENV is not 'development')")
		return
	}

	workDir, _ := os.Getwd()

	// WebSocket test page
	r.Get("/test-socket", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "examples", "test_socket.html"))
	})

	// FCM test page
	r.Get("/test-fcm", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "examples", "test_fcm.html"))
	})

	// Firebase messaging service worker
	r.Get("/firebase-messaging-sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "examples", "firebase-messaging-sw.js"))
	})

	// API: Send test notification
	r.Post("/test/fcm/test", func(w http.ResponseWriter, r *http.Request) {
		if fcmClient == nil {
			http.Error(w, `{"success": false, "message": "FCM client not initialized"}`, http.StatusServiceUnavailable)
			return
		}

		var req struct {
			Token string            `json:"token"`
			Title string            `json:"title"`
			Body  string            `json:"body"`
			Data  map[string]string `json:"data"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"success": false, "message": "Invalid request body: %v"}`, err), http.StatusBadRequest)
			return
		}

		if req.Token == "" {
			http.Error(w, `{"success": false, "message": "Token is required"}`, http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, `{"success": false, "message": "Title is required"}`, http.StatusBadRequest)
			return
		}

		if req.Body == "" {
			http.Error(w, `{"success": false, "message": "Body is required"}`, http.StatusBadRequest)
			return
		}

		// Create notification
		notification := fcm.NewNotificationBuilder().
			SetTitle(req.Title).
			SetBody(req.Body).
			Build()

		// Prepare data
		data := req.Data
		if data == nil {
			data = make(map[string]string)
		}

		// Send notification
		ctx := context.Background()
		messageID, err := fcmClient.SendToToken(ctx, req.Token, notification, data)
		if err != nil {
			logger.Errorf("Failed to send FCM notification: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message_id": messageID,
		})
	})

	// API: Subscribe tokens to topic
	r.Post("/test/fcm/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if fcmClient == nil {
			http.Error(w, `{"success": false, "message": "FCM client not initialized"}`, http.StatusServiceUnavailable)
			return
		}

		var req struct {
			Tokens []string `json:"tokens"`
			Topic  string   `json:"topic"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"success": false, "message": "Invalid request body: %v"}`, err), http.StatusBadRequest)
			return
		}

		if len(req.Tokens) == 0 {
			http.Error(w, `{"success": false, "message": "Tokens array is required"}`, http.StatusBadRequest)
			return
		}

		if req.Topic == "" {
			http.Error(w, `{"success": false, "message": "Topic is required"}`, http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		response, err := fcmClient.SubscribeToTopic(ctx, req.Tokens, req.Topic)
		if err != nil {
			logger.Errorf("Failed to subscribe tokens to topic: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"success_count": response.SuccessCount,
			"failure_count": len(response.Errors),
			"errors":        response.Errors,
		})
	})

	// API: Unsubscribe tokens from topic
	r.Post("/test/fcm/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		if fcmClient == nil {
			http.Error(w, `{"success": false, "message": "FCM client not initialized"}`, http.StatusServiceUnavailable)
			return
		}

		var req struct {
			Tokens []string `json:"tokens"`
			Topic  string   `json:"topic"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"success": false, "message": "Invalid request body: %v"}`, err), http.StatusBadRequest)
			return
		}

		if len(req.Tokens) == 0 {
			http.Error(w, `{"success": false, "message": "Tokens array is required"}`, http.StatusBadRequest)
			return
		}

		if req.Topic == "" {
			http.Error(w, `{"success": false, "message": "Topic is required"}`, http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		response, err := fcmClient.UnsubscribeFromTopic(ctx, req.Tokens, req.Topic)
		if err != nil {
			logger.Errorf("Failed to unsubscribe tokens from topic: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"success_count": response.SuccessCount,
			"failure_count": len(response.Errors),
			"errors":        response.Errors,
		})
	})

	logger.Info("Test pages and FCM test APIs initialized (development mode only)")
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
	logger.Info("WebSocket Endpoint: ws://localhost:3000/ws")

	// Only log test pages in development
	appEnv := utils.GetEnv("APP_ENV", "production")
	if appEnv == "development" {
		logger.Info("WebSocket Test: http://localhost:3000/test-socket")
		logger.Info("FCM Test: http://localhost:3000/test-fcm")
	}

	if err := http.ListenAndServe(":3000", r); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
