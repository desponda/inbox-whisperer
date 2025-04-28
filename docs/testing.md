# Testing Guidelines

## Overview
This document outlines the testing standards and practices for the Inbox Whisperer project. Following these guidelines ensures consistent, reliable, and maintainable tests across the codebase.

## Test Categories

### Unit Tests
- Test individual components in isolation
- Use mocks for external dependencies
- Should be fast and not require external services
- Located alongside the code being tested with `_test.go` suffix

### Integration Tests
- **ALWAYS use testcontainers-go for any tests requiring external services**
- Never rely on external services or local installations
- Each test should get its own isolated container instance
- Common integration test targets:
  - Database operations
  - External API interactions
  - Message queues
  - Cache services

### End-to-End Tests
- Test complete features from user perspective
- Use testcontainers for any required services
- Consider using testing containers for the application itself

## Using Testcontainers

### Standard Setup Pattern
```go
func setupTestContainer(t *testing.T) (*sql.DB, func()) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:         true,
    })
    require.NoError(t, err)

    // Get connection details
    mappedPort, err := container.MappedPort(ctx, "5432")
    require.NoError(t, err)
    hostIP, err := container.Host(ctx)
    require.NoError(t, err)

    // Create connection string and connect
    connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", 
        hostIP, mappedPort.Port())
    db, err := sql.Open("pgx", connStr)
    require.NoError(t, err)
    require.NoError(t, db.Ping())

    // Return cleanup function
    cleanup := func() {
        db.Close()
        container.Terminate(ctx)
    }

    return db, cleanup
}

func TestExample(t *testing.T) {
    db, cleanup := setupTestContainer(t)
    defer cleanup()

    // Run your tests...
}
```

### Best Practices
1. **Always use cleanup functions**
   - Return a cleanup function from setup
   - Use `defer cleanup()` immediately after setup
   - Ensure all resources are properly cleaned up

2. **Container Configuration**
   - Use specific version tags for images
   - Set appropriate environment variables
   - Configure health checks using `WaitingFor`
   - Set resource limits if needed

3. **Test Isolation**
   - Each test should get its own container
   - Don't share containers between tests
   - Clean up all data between test runs

4. **Error Handling**
   - Use `require` for setup operations
   - Handle cleanup errors appropriately
   - Log any cleanup failures

5. **Performance**
   - Use `testing.Short()` to skip integration tests when appropriate
   - Consider parallel test execution with `t.Parallel()`
   - Reuse container images across test runs

### Supported Services
Common testcontainer configurations are available for:
- PostgreSQL
- Redis
- MongoDB
- RabbitMQ
- Elasticsearch
- And more...

## Running Tests
```bash
# Run all tests
go test ./...

# Run only unit tests
go test -short ./...

# Run specific integration tests
go test ./internal/data/... -v

# Run with race detection
go test -race ./...
```

## CI/CD Considerations
- Integration tests using testcontainers work seamlessly in CI environments
- No need for service setup in CI pipelines
- Consider resource limits and cleanup in CI environments
- Use appropriate timeouts for container operations

## Adding New Integration Tests
When adding new integration tests:
1. Always use testcontainers for external service dependencies
2. Follow the standard setup pattern shown above
3. Ensure proper cleanup
4. Document any special container configuration
5. Consider adding the configuration to shared test utilities 