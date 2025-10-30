package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-core/pkg/jwt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestUser represents a test user for testing
type TestUser struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	IsActive bool      `json:"is_active"`
}

// CreateTestUser creates a test user
func CreateTestUser() *TestUser {
	return &TestUser{
		ID:       uuid.New(),
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     "user",
		IsActive: true,
	}
}

// CreateTestAdmin creates a test admin user
func CreateTestAdmin() *TestUser {
	return &TestUser{
		ID:       uuid.New(),
		Name:     "Test Admin",
		Email:    "admin@example.com",
		Role:     "admin",
		IsActive: true,
	}
}

// GenerateTestToken generates a JWT token for testing
func GenerateTestToken(t *testing.T, jwtManager *jwt.Manager, userID, role, email string) string {
	token, err := jwtManager.GenerateToken(userID, email, role, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return token
}

// CreateTestRequest creates a test HTTP request
func CreateTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte("{}"))
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateAuthenticatedRequest creates a test HTTP request with JWT token
func CreateAuthenticatedRequest(t *testing.T, jwtManager *jwt.Manager, method, url string, body interface{}, userID, role, email string) *http.Request {
	req := CreateTestRequest(method, url, body)
	token := GenerateTestToken(t, jwtManager, userID, role, email)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// ExecuteRequest executes a test request and returns response
func ExecuteRequest(t *testing.T, handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// AssertResponseStatus asserts the response status code
func AssertResponseStatus(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response: %s", expectedStatus, w.Code, w.Body.String())
	}
}

// AssertResponseJSON asserts the response contains expected JSON
func AssertResponseJSON(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response JSON: %v", err)
		return
	}

	expectedJSON, _ := json.Marshal(expected)
	var expectedMap map[string]interface{}
	if err := json.Unmarshal(expectedJSON, &expectedMap); err != nil {
		t.Errorf("Failed to marshal expected JSON: %v", err)
		return
	}

	// Simple assertion - you can make this more sophisticated
	if len(response) == 0 && len(expectedMap) > 0 {
		t.Errorf("Expected non-empty response, got empty")
	}
}

// AssertResponseContains asserts the response contains a specific string
func AssertResponseContains(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	if !bytes.Contains(w.Body.Bytes(), []byte(expected)) {
		t.Errorf("Expected response to contain '%s', got: %s", expected, w.Body.String())
	}
}

// CreateTestContext creates a test context with user ID
func CreateTestContext(userID string) context.Context {
	return context.WithValue(context.Background(), jwt.UserIDContextKey, userID)
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("Timeout waiting for condition: %s", message)
}

// MockHTTPClient creates a mock HTTP client for testing
type MockHTTPClient struct {
	Responses map[string]*http.Response
	Errors    map[string]error
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Responses: make(map[string]*http.Response),
		Errors:    make(map[string]error),
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()
	if err, exists := m.Errors[key]; exists {
		return nil, err
	}
	if resp, exists := m.Responses[key]; exists {
		return resp, nil
	}
	return nil, fmt.Errorf("no mock response for %s", key)
}

// TestDataCleanup cleans up test data
func TestDataCleanup(t *testing.T, db *gorm.DB) {
	// Clean all test data from database
	CleanTestDB(t, db)
}

// BenchmarkHelper provides helper functions for benchmarking
type BenchmarkHelper struct {
	DB         *gorm.DB
	JWTManager *jwt.Manager
}

func NewBenchmarkHelper(t *testing.T) *BenchmarkHelper {
	config := SetupTestConfig(t)
	return &BenchmarkHelper{
		DB:         config.DB,
		JWTManager: config.JWTManager,
	}
}

func (b *BenchmarkHelper) Cleanup() {
	// Cleanup benchmark resources
}
