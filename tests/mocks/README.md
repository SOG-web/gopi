# Test Mocks Directory Structure

This directory contains modular test mocks organized by module for better maintainability and organization.

## Directory Structure

```
tests/mocks/
├── README.md              # This file
├── chat/
│   └── mocks.go          # Chat module mocks (Group, Message, ChatService)
├── post/
│   └── mocks.go          # Post module mocks (Post, Comment, PostService)
└── user/
    └── mocks.go          # User module mocks (User, EmailService, UserService)
```

## Organization Benefits

### ✅ **Better Organization**

- Mocks are grouped by module/domain
- Easy to find relevant mocks for each module
- Clear separation of concerns

### ✅ **Maintainability**

- Smaller, focused mock files
- Easier to update when interfaces change
- Module-specific mock implementations

### ✅ **Reusability**

- Mocks can be imported by module: `moduleMocks "gopi.com/tests/mocks/chat"`
- Shared across all test files within the module
- Consistent mock behavior

### ✅ **Scalability**

- New modules can have their own mock directories
- Easy to add module-specific mocks
- No monolithic mock file

## Usage Examples

### Chat Module

```go
import (
    chatMocks "gopi.com/tests/mocks/chat"
    "gopi.com/internal/app/chat"
    chatModel "gopi.com/internal/domain/chat/model"
)

// Create mocks
mockGroupRepo := new(chatMocks.MockGroupRepository)
mockMessageRepo := new(chatMocks.MockMessageRepository)
mockChatService := new(chatMocks.MockChatService)

// Use in tests
chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)
```

### Post Module

```go
import (
    postMocks "gopi.com/tests/mocks/post"
    "gopi.com/internal/app/post"
    postModel "gopi.com/internal/domain/post/model"
)

// Create mocks
mockPostRepo := new(postMocks.MockPostRepository)
mockCommentRepo := new(postMocks.MockCommentRepository)
mockPostService := new(postMocks.MockPostService)

// Use in tests
postService := post.NewPostService(mockPostRepo, mockCommentRepo)
```

### User Module

```go
import (
    userMocks "gopi.com/tests/mocks/user"
    "gopi.com/internal/app/user"
    userModel "gopi.com/internal/domain/user/model"
)

// Create mocks
mockUserRepo := new(userMocks.MockUserRepository)
mockEmailService := new(userMocks.MockEmailService)
mockUserService := new(userMocks.MockUserService)

// Use in tests
userService := user.NewUserService(mockUserRepo, mockEmailService)
```

## Available Mocks

### Chat Module (`/chat/mocks.go`)

- `MockGroupRepository` - Chat group repository interface
- `MockMessageRepository` - Chat message repository interface
- `MockChatService` - Chat service interface

### Post Module (`/post/mocks.go`)

- `MockPostRepository` - Post repository interface
- `MockCommentRepository` - Comment repository interface
- `MockPostService` - Post service interface

### User Module (`/user/mocks.go`)

- `MockUserRepository` - User repository interface
- `MockEmailService` - Email service interface
- `MockUserService` - User service interface

## Migration from Shared Mocks

If you have existing tests using the old shared mocks structure:

### Old (Shared Mocks)

```go
import "gopi.com/tests/mocks"

// ❌ Old way
mockGroupRepo := new(mocks.MockGroupRepository)
mockChatService := new(mocks.MockChatService)
```

### New (Modular Mocks)

```go
import chatMocks "gopi.com/tests/mocks/chat"

// ✅ New way
mockGroupRepo := new(chatMocks.MockGroupRepository)
mockChatService := new(chatMocks.MockChatService)
```

## Best Practices

1. **Import by Module**: Use module-specific imports for clarity
2. **Consistent Naming**: Follow the `Mock[TypeName]` naming convention
3. **Complete Implementation**: Ensure all interface methods are mocked
4. **Documentation**: Add comments explaining complex mock behaviors
5. **Testing**: Test mock behavior in addition to business logic

## Adding New Mocks

When adding mocks for a new module:

1. Create a new directory: `mkdir tests/mocks/newmodule`
2. Create `mocks.go` file with mock implementations
3. Follow the established patterns and naming conventions
4. Update this README with the new module information

This modular approach ensures that our test infrastructure remains maintainable and scalable as the application grows!
