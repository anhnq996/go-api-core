# Testing Guide for ApiCore

This document provides comprehensive information about testing in the ApiCore application.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Test Configuration](#test-configuration)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

ApiCore uses a comprehensive testing strategy that includes:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **Mock Objects**: For external dependencies
- **Test Utilities**: Helper functions and test data
- **Coverage Reports**: Track test coverage

## Test Structure

```
├── test/                          # Test utilities and integration tests
│   ├── config.go                 # Test configuration setup
│   ├── utils.go                  # Test utilities and helpers
│   └── integration_test.go       # Integration tests
├── internal/
│   ├── repositories/
│   │   └── *_test.go            # Repository unit tests
│   ├── app/
│   │   └── user/
│   │       ├── service_test.go   # Service unit tests
│   │       └── handler_test.go   # Handler unit tests
├── pkg/
│   └── cache/
│       └── mock.go               # Mock implementations
├── scripts/
│   └── test.sh                  # Test runner script
├── Makefile                      # Development commands
└── test.env                      # Test environment variables
```

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage

# Run tests in watch mode
make test-watch
```

### Using Test Script

```bash
# Run all tests
./scripts/test.sh all

# Run unit tests
./scripts/test.sh unit

# Run integration tests
./scripts/test.sh integration

# Run with coverage
./scripts/test.sh all coverage

# Verbose output
./scripts/test.sh all -v
```

### Using Go Commands Directly

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/repositories/...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

## Writing Tests

### Test File Naming

- Unit tests: `*_test.go` in the same package
- Integration tests: `*_test.go` in the `test/` directory
- Test files should be in the same package as the code being tested

### Basic Test Structure

```go
package repository_test

import (
    "testing"
    "api-core/test"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
    // Setup
    config := test.SetupTestConfig(t)
    defer test.CleanupTestConfig(t, config)

    // Test implementation
    // ...

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Test Utilities

The `test` package provides utilities for:

- **SetupTestConfig**: Initialize test configuration
- **CreateTestUser**: Create test user data
- **CreateAuthenticatedRequest**: Create HTTP requests with JWT
- **ExecuteRequest**: Execute HTTP requests
- **AssertResponseStatus**: Assert HTTP response status
- **AssertResponseJSON**: Assert JSON response content

### Mock Objects

Use mock objects for external dependencies:

```go
// Mock cache
mockCache := cache.NewMockCache()

// Mock HTTP client
mockClient := test.NewMockHTTPClient()
```

## Test Configuration

### Environment Variables

Test configuration is loaded from `test.env`:

```bash
# Test Mode
GO_ENV=test
TEST_MODE=true

# Database (SQLite in-memory)
DB_DRIVER=sqlite
DB_HOST=:memory:

# JWT Configuration
JWT_SECRET_KEY=test-secret-key-min-32-chars-long-for-testing
```

### Test Database

Tests use SQLite in-memory database for:

- Fast test execution
- No external dependencies
- Automatic cleanup
- Isolation between tests

### Test Data

- Use `test.CreateTestUser()` for consistent test data
- Clean up test data in `defer` statements
- Use unique identifiers to avoid conflicts

## Best Practices

### 1. Test Organization

```go
func TestUserService_Create(t *testing.T) {
    // Arrange (Setup)
    config := test.SetupTestConfig(t)
    defer test.CleanupTestConfig(t, config)

    // Act (Execute)
    result, err := service.Create(ctx, user)

    // Assert (Verify)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 2. Test Naming

- Use descriptive test names: `TestUserRepository_Create_WithValidData`
- Include the method being tested
- Describe the scenario or condition

### 3. Test Isolation

- Each test should be independent
- Use `t.Parallel()` for parallel execution when safe
- Clean up resources in `defer` statements

### 4. Assertions

- Use `require` for critical assertions that should stop the test
- Use `assert` for non-critical assertions
- Provide meaningful error messages

### 5. Test Data

- Use factories for creating test data
- Keep test data minimal and focused
- Use realistic but simple data

### 6. Error Testing

```go
func TestUserService_Create_WithInvalidData(t *testing.T) {
    // Test error conditions
    _, err := service.Create(ctx, invalidUser)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
}
```

## Coverage

### Running Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# View coverage in terminal
go tool cover -func=coverage.out
```

### Coverage Goals

- **Unit Tests**: Aim for 80%+ coverage
- **Critical Paths**: Aim for 90%+ coverage
- **Integration Tests**: Focus on happy paths and error cases

## Troubleshooting

### Common Issues

1. **Database Connection Errors**

   ```bash
   # Ensure test environment is set
   export GO_ENV=test
   ```

2. **Import Cycle Errors**

   ```bash
   # Use test package for test utilities
   import "api-core/test"
   ```

3. **Mock Not Working**
   ```bash
   # Ensure mock implements the interface correctly
   var _ cache.Cache = (*cache.MockCache)(nil)
   ```

### Debug Tests

```bash
# Run specific test with verbose output
go test -v -run TestUserRepository_Create ./internal/repositories/...

# Run tests with race detection
go test -race ./...

# Run tests with profiling
go test -cpuprofile=cpu.prof ./...
```

### Performance Testing

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -memprofile=mem.prof ./...
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: make ci-test
```

### Pre-commit Hooks

```bash
# Install pre-commit hook
echo '#!/bin/bash\nmake test' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/effective_go.html#testing)
- [Mock Testing in Go](https://golang.org/doc/effective_go.html#interfaces)
