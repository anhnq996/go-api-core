package main

import (
	"net/http"
	"os"
	"path/filepath"

	"anhnq/api-core/config"
	"anhnq/api-core/internal/routes"
	"anhnq/api-core/internal/wire"
	"anhnq/api-core/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Khởi tạo logger
	if err := logger.Init(logger.Config{
		Level:        "debug",                 // debug, info, warn, error
		Output:       "console,file",          // console, file, loki (có thể kết hợp)
		FilePath:     "storages/logs/app.log", // file log path
		LokiURL:      "http://localhost:3100", // Loki server URL
		EnableCaller: false,                   // hiển thị file:line
		PrettyPrint:  true,                    // format đẹp cho console
	}); err != nil {
		panic(err)
	}

	logger.Info("Starting ApiCore application...")

	// Kết nối database
	dbConfig := config.GetDefaultDatabaseConfig()
	db, err := config.ConnectDatabase(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Database connected successfully")

	// Wire tự động khởi tạo tất cả dependencies
	controllers := wire.InitializeApp(db)
	logger.Info("Dependencies initialized successfully")

	// Khởi tạo router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID) // Tạo unique ID cho mỗi request
	r.Use(middleware.Recoverer) // Recover từ panic
	r.Use(logger.Middleware())  // Log requests/responses với đầy đủ thông tin

	// Health check endpoint
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// Documentation routes
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

	// Đăng ký tất cả routes
	routes.RegisterRoutes(r, controllers)

	// Start server
	logger.Info("Server starting on :3000")
	logger.Info("Documentation: http://localhost:3000/docs")
	logger.Info("Swagger UI: http://localhost:3000/swagger")

	if err := http.ListenAndServe(":3000", r); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
