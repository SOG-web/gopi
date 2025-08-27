# GoPi Backend Tests

This directory contains comprehensive tests for the GoPi backend API, focusing on the chat module that was converted from Python to Go.

## Test Structure

```
tests/
├── README.md              # This file
├── chat/                  # Chat module tests
│   ├── chat_handler_test.go    # HTTP handler tests
│   ├── chat_service_test.go    # Business logic tests
│   ├── chat_repo_test.go       # Database repository tests
│   └── chat_websocket_test.go  # WebSocket functionality tests
└── integration/           # Integration tests (future)
```

## Test Coverage

### 1. Chat Handler Tests (`chat_handler_test.go`)
Tests the HTTP REST API endpoints for chat functionality:

- **Group Management**:
  - `POST /chat/groups` - Create new chat groups
  - `GET /chat/groups` - List user's groups with pagination
  - `GET /chat/groups/{slug}` - Get specific group details
  - `PUT /chat/groups/{slug}` - Update group (creator only)
  - `DELETE /chat/groups/{slug}` - Delete group (creator only)

- **Group Membership**:
  - `POST /chat/groups/{slug}/join` - Join a group
  - `POST /chat/groups/{slug}/leave` - Leave a group

- **Validation & Security**:
  - Group name length limits (20 characters max)
  - Authentication requirements
  - Authorization checks (creator vs member permissions)
  - Group membership validation

### 2. Chat Service Tests (`chat_service_test.go`)
Tests the business logic layer with comprehensive scenarios:

- **Group Operations**:
  - Creating groups with automatic creator inclusion
  - Retrieving groups by ID and slug
  - Updating groups with validation
  - Deleting groups with permission checks
  - Searching groups by name

- **Membership Management**:
  - Adding members (creator and existing member permissions)
  - Removing members (self-removal and creator permissions)
  - Preventing duplicate memberships
  - Authorization validation

- **Message Handling**:
  - Sending messages with membership validation
  - Message length limits (1000 characters max)
  - Retrieving message history with pagination
  - Message ownership validation

- **Business Rules**:
  - Creator automatically becomes first member
  - Only group members can send messages
  - Only creators or message senders can delete messages
  - Group name length enforcement

### 3. Chat Repository Tests (`chat_repo_test.go`)
Tests database operations using SQLite in-memory database:

- **CRUD Operations**:
  - Creating groups and messages
  - Reading by ID, slug, and member ID
  - Updating records
  - Deleting records

- **Advanced Queries**:
  - Searching groups by name (case-insensitive)
  - Retrieving groups by member ID
  - Pagination support
  - Ordering by creation time

- **Data Integrity**:
  - JSON serialization/deserialization of member arrays
  - Foreign key relationships
  - Unique constraints (slugs)
  - Auto-incrementing IDs

### 4. WebSocket Tests (`chat_websocket_test.go`)
Tests real-time messaging functionality:

- **Connection Management**:
  - WebSocket upgrade handling
  - Authentication validation
  - Group membership verification

- **Message Handling**:
  - Chat message broadcasting
  - Typing indicators
  - Message format validation
  - JSON serialization/deserialization

- **Security**:
  - Unauthorized access prevention
  - Group membership validation
  - Message content validation

- **Real-time Features**:
  - Connection cleanup
  - Group-based message isolation
  - Concurrent connection handling

## Running Tests

### Run All Tests
```bash
# From the project root
cd /Users/rou/Desktop/bgw/gopi
go test ./tests/...
```

### Run Specific Test Suites
```bash
# Chat handler tests only
go test ./tests/chat -run TestChatHandler

# Chat service tests only
go test ./tests/chat -run TestChatService

# Chat repository tests only
go test ./tests/chat -run TestGormGroupRepository

# WebSocket tests only
go test ./tests/chat -run TestWebSocket
```

### Run Tests with Coverage
```bash
# Generate coverage report
go test ./tests/... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# Show coverage summary
go tool cover -func=coverage.out
```

### Run Tests with Verbose Output
```bash
go test ./tests/... -v
```

### Run Tests in Parallel
```bash
go test ./tests/... -parallel=4
```

## Test Dependencies

The tests use the following packages:
- `github.com/stretchr/testify` - Assertions and mocking
- `github.com/gin-gonic/gin` - HTTP testing
- `gorm.io/driver/sqlite` - In-memory database for testing
- `github.com/gorilla/websocket` - WebSocket testing

## Test Database

Tests use SQLite in-memory database for:
- Fast test execution
- Isolated test environments
- No external database dependencies
- Automatic cleanup after each test

## Mocking Strategy

Tests use the `testify/mock` package to:
- Mock external dependencies (services, repositories)
- Control test scenarios
- Verify interaction patterns
- Isolate unit tests from external systems

## Continuous Integration

These tests are designed to run in CI/CD pipelines and provide:
- Fast feedback on code changes
- Regression prevention
- Documentation of expected behavior
- Confidence in deployments

## Test Categories

### Unit Tests
- Test individual functions/methods
- Mock external dependencies
- Fast execution
- Located in `*_test.go` files

### Integration Tests
- Test component interactions
- Use real database connections
- Slower execution
- Located in `integration/` directory (planned)

## Best Practices Followed

1. **Descriptive Test Names**: Each test clearly describes what it validates
2. **Table-Driven Tests**: Multiple test cases in a single test function
3. **Proper Setup/Teardown**: Each test is isolated and independent
4. **Mock Verification**: All mock expectations are verified
5. **Error Case Testing**: Both success and failure scenarios tested
6. **Edge Case Coverage**: Boundary conditions and unusual inputs tested
7. **Security Testing**: Authentication and authorization thoroughly tested

## Adding New Tests

When adding new functionality:

1. Create tests alongside the code (TDD approach recommended)
2. Follow the existing naming conventions
3. Include both positive and negative test cases
4. Add tests for security and validation
5. Update this README if new test patterns are introduced

## Troubleshooting

### Common Issues

1. **Mock Setup Errors**: Ensure all mock expectations are properly configured
2. **Database Errors**: Check that migrations are applied correctly
3. **Import Errors**: Verify all dependencies are installed with `go mod tidy`
4. **Race Conditions**: WebSocket tests may need timing adjustments

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
- Performance/load testing
- End-to-end API testing
- Browser-based WebSocket testing
- Chaos engineering tests
