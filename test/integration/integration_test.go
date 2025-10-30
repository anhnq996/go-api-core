package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-core/internal/app/user"
	repository "api-core/internal/repositories"
	"api-core/internal/routes"
	"api-core/pkg/jwt"
	"api-core/test"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAPI_Integration(t *testing.T) {
	// Setup with PostgreSQL container (auto migrate + seed)
	config := test.SetupTestContainerConfig(t, true, true)
	defer test.CleanupTestContainerConfig(t, config)

	// Setup dependencies
	userRepo := repository.NewUserRepository(config.DB)
	userService := user.NewService(userRepo, config.Cache, config.Storage)
	userHandler := user.NewHandler(userService)
	jwtManager := config.JWTManager
	jwtBlacklist := jwt.NewBlacklist(config.Cache)

	// Setup routes
	r := chi.NewRouter()
	controllers := &routes.Controllers{
		UserHandler:  userHandler,
		JWTManager:   jwtManager,
		JWTBlacklist: jwtBlacklist,
		Cache:        config.Cache,
	}
	routes.RegisterRoutes(r, controllers)

	// Test 1: Create User
	t.Run("Create User", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Integration Test User",
			"email": "integration@example.com",
		}

		req := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	// Test 2: Get Users List
	t.Run("Get Users List", func(t *testing.T) {
		req := test.CreateAuthenticatedRequest(t, jwtManager, "GET", "/api/v1/users", nil, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	// Test 3: Get Specific User
	t.Run("Get Specific User", func(t *testing.T) {
		// First create a user to get its ID
		userData := map[string]interface{}{
			"name":  "Test User for Get",
			"email": "testget@example.com",
		}

		createReq := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		data := createResponse["data"].(map[string]interface{})
		userID := data["id"].(string)

		// Now get the user
		req := test.CreateAuthenticatedRequest(t, jwtManager, "GET", "/api/v1/users/"+userID, nil, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	// Test 4: Update User
	t.Run("Update User", func(t *testing.T) {
		// First create a user
		userData := map[string]interface{}{
			"name":  "Test User for Update",
			"email": "testupdate@example.com",
		}

		createReq := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		data := createResponse["data"].(map[string]interface{})
		userID := data["id"].(string)

		// Update the user
		updateData := map[string]interface{}{
			"name":  "Updated Integration User",
			"email": "updated@example.com",
		}

		req := test.CreateAuthenticatedRequest(t, jwtManager, "PUT", "/api/v1/users/"+userID, updateData, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	// Test 5: Delete User
	t.Run("Delete User", func(t *testing.T) {
		// First create a user
		userData := map[string]interface{}{
			"name":  "Test User for Delete",
			"email": "testdelete@example.com",
		}

		createReq := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)

		data := createResponse["data"].(map[string]interface{})
		userID := data["id"].(string)

		// Delete the user
		req := test.CreateAuthenticatedRequest(t, jwtManager, "DELETE", "/api/v1/users/"+userID, nil, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

func TestUserAPI_Authentication(t *testing.T) {
	// Setup
	config := test.SetupTestContainerConfig(t, true, true)
	defer test.CleanupTestContainerConfig(t, config)

	// Setup dependencies
	userRepo := repository.NewUserRepository(config.DB)
	userService := user.NewService(userRepo, config.Cache, config.Storage)
	userHandler := user.NewHandler(userService)
	jwtManager := config.JWTManager
	jwtBlacklist := jwt.NewBlacklist(config.Cache)

	// Setup routes
	r := chi.NewRouter()
	controllers := &routes.Controllers{
		UserHandler:  userHandler,
		JWTManager:   jwtManager,
		JWTBlacklist: jwtBlacklist,
		Cache:        config.Cache,
	}
	routes.RegisterRoutes(r, controllers)

	// Test 1: Unauthorized Request
	t.Run("Unauthorized Request", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test 2: Invalid Token
	t.Run("Invalid Token", func(t *testing.T) {
		req := test.CreateTestRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test 3: Valid Token
	t.Run("Valid Token", func(t *testing.T) {
		req := test.CreateAuthenticatedRequest(t, jwtManager, "GET", "/api/v1/users", nil, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserAPI_Validation(t *testing.T) {
	// Setup
	config := test.SetupTestContainerConfig(t, true, true)
	defer test.CleanupTestContainerConfig(t, config)

	// Setup dependencies
	userRepo := repository.NewUserRepository(config.DB)
	userService := user.NewService(userRepo, config.Cache, config.Storage)
	userHandler := user.NewHandler(userService)
	jwtManager := config.JWTManager
	jwtBlacklist := jwt.NewBlacklist(config.Cache)

	// Setup routes
	r := chi.NewRouter()
	controllers := &routes.Controllers{
		UserHandler:  userHandler,
		JWTManager:   jwtManager,
		JWTBlacklist: jwtBlacklist,
		Cache:        config.Cache,
	}
	routes.RegisterRoutes(r, controllers)

	// Test 1: Invalid Email Format
	t.Run("Invalid Email Format", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Test User",
			"email": "invalid-email",
		}

		req := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 2: Missing Required Fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		userData := map[string]interface{}{
			"name": "Test User",
			// Missing email
		}

		req := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test 3: Duplicate Email
	t.Run("Duplicate Email", func(t *testing.T) {
		// First create a user
		userData := map[string]interface{}{
			"name":  "First User",
			"email": "duplicate@example.com",
		}

		createReq := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", userData, "admin-123", "admin", "admin@example.com")
		createW := httptest.NewRecorder()
		r.ServeHTTP(createW, createReq)

		// Try to create another user with same email
		duplicateData := map[string]interface{}{
			"name":  "Second User",
			"email": "duplicate@example.com",
		}

		req := test.CreateAuthenticatedRequest(t, jwtManager, "POST", "/api/v1/users", duplicateData, "admin-123", "admin", "admin@example.com")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
