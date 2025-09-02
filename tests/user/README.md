# User Module Tests

This directory contains comprehensive tests for the GoPi backend user module, covering authentication, user management, and related functionality.

## Test Structure

```
tests/user/
├── README.md                    # This file
├── user_handler_test.go         # HTTP handler/API tests for user operations
├── user_service_test.go         # Business logic tests
├── user_repo_test.go            # Database repository tests
├── auth_handler_test.go         # HTTP handler/API tests for authentication
└── ../mocks/user/mocks.go       # Mock implementations
```

## Test Coverage

### 1. User Handler Tests (`user_handler_test.go`)

Tests the HTTP REST API endpoints for user management:

- **Profile Management**:

  - `GET /users/profile` - Get current user's profile
  - `PUT /users/profile` - Update current user's profile

- **Admin Operations**:

  - `GET /users` - Get all users (admin only)
  - `GET /users/staff` - Get all staff users (admin only)
  - `GET /users/verified` - Get all verified users (admin only)
  - `GET /users/unverified` - Get all unverified users (admin only)
  - `GET /users/:id` - Get specific user by ID (admin only)

- **Security & Validation**:
  - Authentication middleware integration
  - Admin permission checks
  - Input validation
  - Error handling

### 2. Auth Handler Tests (`auth_handler_test.go`)

Tests the HTTP REST API endpoints for authentication:

- **User Registration**:

  - `POST /api/auth/register/` - User registration with validation
  - Email/username uniqueness checks
  - JWT token generation
  - OTP email sending

- **User Login**:

  - `POST /api/auth/login/` - User authentication
  - Password verification
  - Account status validation
  - JWT token generation

- **User Logout**:

  - `GET /api/auth/logout/` - User logout
  - JWT token blacklisting
  - Authentication middleware

- **OTP Verification**:

  - `POST /api/auth/verify/` - Email verification
  - OTP validation
  - Account activation

- **Password Management**:

  - `PUT /api/auth/change-password/` - Password change
  - Current password verification
  - New password validation

- **Account Management**:

  - `DELETE /api/auth/delete/` - Account deletion
  - User data removal

- **OTP Resend**:
  - `PUT /api/auth/resend-otp/:id/` - OTP resend
  - User lookup and validation

### 3. User Service Tests (`user_service_test.go`)

Tests the business logic layer with comprehensive scenarios:

- **User Registration**:

  - Email and username uniqueness validation
  - Password hashing
  - OTP generation and email sending
  - User creation with proper defaults

- **Authentication**:

  - User login with email/password
  - Password verification
  - User verification status checks
  - Account activation checks
  - Last login timestamp updates

- **OTP Verification**:

  - OTP validation
  - User verification status updates
  - Welcome email sending

- **Password Management**:

  - Password change functionality
  - Old password verification
  - Password hashing and updates

- **User Management**:

  - User profile updates
  - Account deletion
  - User retrieval by ID, email, username
  - User listing with pagination

- **Admin Functions**:

  - Staff user management
  - User verification management
  - User statistics
  - Bulk operations

- **Utility Functions**:
  - Email validation
  - Username validation
  - Password hashing and verification

### 4. User Repository Tests (`user_repo_test.go`)

Tests database operations using SQLite in-memory database:

- **CRUD Operations**:

  - User creation with all fields
  - User retrieval by ID, email, username
  - User updates
  - User deletion

- **Authentication Operations**:

  - Email/password authentication
  - OTP-based authentication
  - Password updates
  - OTP updates
  - Verification status updates
  - Last login updates

- **Admin Queries**:

  - Get all users
  - Get staff users
  - Get verified/unverified users
  - User listing with pagination

- **Validation Helpers**:

  - Email existence checks
  - Username existence checks

- **Data Integrity**:
  - Foreign key relationships
  - Unique constraints
  - Default values
  - Auto-incrementing fields

## Running Tests

### Run All User Tests

```bash
# From the project root
cd /Users/rou/Desktop/bgw/gopi
go test ./tests/user -v
```

### Run Specific Test Files

```bash
# User handler tests only
go test ./tests/user -run TestUserHandler -v

# Auth handler tests only
go test ./tests/user -run TestAuthHandler -v

# Service tests only
go test ./tests/user -run TestUserService -v

# Repository tests only
go test ./tests/user -run TestUserRepositoryGORM -v
```

### Run Tests with Coverage

```bash
# Generate coverage report for user tests
go test ./tests/user -coverprofile=user_coverage.out -v

# View coverage in browser
go tool cover -html=user_coverage.out

# Show coverage summary
go tool cover -func=user_coverage.out
```

### Run Tests in Parallel

```bash
go test ./tests/user -parallel=4 -v
```

## Test Dependencies

The tests use the following packages:

- `github.com/stretchr/testify` - Assertions and mocking
- `github.com/gin-gonic/gin` - HTTP testing
- `gorm.io/driver/sqlite` - In-memory database for testing
- `golang.org/x/crypto/bcrypt` - Password hashing

## Test Database

Tests use SQLite in-memory database for:

- Fast test execution
- Isolated test environments
- No external database dependencies
- Automatic cleanup after each test

## Mocking Strategy

Tests use the `testify/mock` package to:

- Mock external dependencies (email service, repositories, JWT service)
- Control test scenarios
- Verify interaction patterns
- Isolate unit tests from external systems

### Mock Services Used:

- **MockUserRepository**: Mock user data operations
- **MockEmailService**: Mock email sending operations
- **MockJWTService**: Mock JWT token operations

## Key Test Patterns

### 1. Table-Driven Tests

All tests use table-driven approach for multiple test cases:

```go
tests := []struct {
    name           string
    requestBody    dto.RegistrationRequest
    expectedStatus int
    mockSetup      func()
}{
    // test cases...
}
```

### 2. Mock Setup and Verification

Each test properly sets up mocks and verifies expectations:

```go
mockUserRepo.On("GetByID", "user123").Return(expectedUser, nil)
// ... test logic ...
mockUserRepo.AssertExpectations(t)
```

### 3. Authentication Testing

Special attention to authentication and authorization:

```go
// Mock auth middleware
auth := router.Group("/api/auth")
auth.Use(func(c *gin.Context) {
    c.Set("user_id", "test-user-id")
    c.Next()
})
```

### 4. Error Testing

Comprehensive error scenarios are tested:

- Repository errors
- Validation errors
- Authentication errors
- Authorization errors
- Network/database errors

### 5. Edge Cases

Tests cover various edge cases:

- Empty inputs
- Invalid data
- Non-existent resources
- Permission denied scenarios
- Token validation failures

## Authentication & Authorization Testing

The tests thoroughly cover authentication and authorization:

- **JWT Token Management**: Token generation, validation, blacklisting
- **User Permissions**: Regular user vs admin access control
- **Session Management**: Login/logout state management
- **OTP Verification**: Email verification flow
- **Password Security**: Secure password handling and validation

## Business Logic Validation

Service layer tests validate complex business rules:

- **Registration Logic**: Email/username uniqueness, password hashing
- **Login Logic**: Password verification, account status checks
- **OTP Logic**: Verification flow, email sending
- **Admin Logic**: Permission checks, bulk operations

## Database Integration Testing

Repository tests ensure data persistence and retrieval:

- **CRUD Operations**: Full lifecycle testing
- **Relationships**: Foreign key constraints
- **Constraints**: Unique constraints, required fields
- **Queries**: Complex filtering and pagination

## Security Testing

Special focus on security aspects:

- **Password Security**: Hashing, verification, strength requirements
- **Authentication Security**: Token validation, session management
- **Authorization Security**: Role-based access control
- **Input Validation**: SQL injection prevention, XSS protection
- **Data Privacy**: Sensitive data handling

## Best Practices Followed

1. **Descriptive Test Names**: Each test clearly describes what it validates
2. **Table-Driven Tests**: Multiple test cases in a single test function
3. **Proper Setup/Teardown**: Each test is isolated and independent
4. **Mock Verification**: All mock expectations are verified
5. **Error Case Testing**: Both success and failure scenarios tested
6. **Edge Case Coverage**: Boundary conditions and unusual inputs tested
7. **Security Testing**: Authentication and authorization thoroughly tested

## Adding New Tests

When adding new user functionality:

1. Follow the existing naming conventions (`Test{Component}_{Functionality}`)
2. Include both positive and negative test cases
3. Add tests for security and validation
4. Update mocks if new interfaces are added
5. Update this README with new test patterns

## Troubleshooting

### Common Issues

1. **Mock Setup Errors**: Ensure all mock expectations are properly configured
2. **Database Errors**: Check that migrations are applied correctly in setup
3. **Import Errors**: Verify all dependencies are installed with `go mod tidy`
4. **Authentication Errors**: Ensure mock middleware is properly configured

### Debugging Tips

1. Use `t.Log()` or `t.Logf()` for debugging output
2. Run tests with `-v` flag for verbose output
3. Use `t.Skip()` to temporarily disable problematic tests
4. Check test isolation - one test shouldn't affect another

## Performance Considerations

- Tests use in-memory database for speed
- Parallel test execution is supported
- Mocking reduces external dependencies
- Tests are designed to run quickly in CI/CD

## Future Enhancements

Planned test improvements:

- Integration tests with real database
- End-to-end API testing
- Load testing for authentication endpoints
- Security vulnerability testing
- Performance benchmarking
- API contract testing with OpenAPI specs

## Test Coverage Summary

| Component       | Test File              | Coverage                           |
| --------------- | ---------------------- | ---------------------------------- |
| User Handler    | `user_handler_test.go` | Profile, Admin operations          |
| Auth Handler    | `auth_handler_test.go` | Registration, Login, OTP, Password |
| User Service    | `user_service_test.go` | Business logic, validation         |
| User Repository | `user_repo_test.go`    | Database operations, queries       |

This comprehensive test suite ensures reliable user management and authentication functionality with proper security, validation, and error handling.
