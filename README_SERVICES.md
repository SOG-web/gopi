# Service Configuration Guide

This application supports both Redis and Database implementations for JWT token blacklisting and password reset functionality. You can choose which implementation to use via environment variables.

## Environment Variables

### JWT Service Configuration
- `USE_DATABASE_JWT=false` (default: false)
  - `false`: Use Redis for JWT token blacklisting (requires Redis)
  - `true`: Use Database for JWT token blacklisting

### Password Reset Service Configuration
- `USE_DATABASE_PWRESET=false` (default: false)
  - `false`: Use Redis for password reset tokens (requires Redis)
  - `true`: Use Database for password reset tokens

## Usage Examples

### Using Redis-based Services (Default)
```bash
# No environment variables needed (defaults to Redis)
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=""
export REDIS_DB=0
```

### Using Database-based Services
```bash
export USE_DATABASE_JWT=true
export USE_DATABASE_PWRESET=true
# No Redis configuration needed
```

### Mixed Configuration
```bash
# Use database for JWT, Redis for password reset
export USE_DATABASE_JWT=true
export USE_DATABASE_PWRESET=false
export REDIS_ADDR=localhost:6379
```

## Service Implementation Details

### JWT Service
- **Redis Implementation**: Uses Redis with TTL for automatic token expiration
- **Database Implementation**: Uses GORM with manual cleanup of expired tokens
- Both implementations provide the same `JWTServiceInterface`

### Password Reset Service
- **Redis Implementation**: Uses Redis with TTL for automatic token expiration
- **Database Implementation**: Uses GORM with manual cleanup and audit trail
- Both implementations provide the same `PasswordResetServiceInterface`

## Database Tables

When using database implementations, the following tables are automatically created:

### JWT Token Blacklisting
```sql
CREATE TABLE blacklisted_tokens (
    id VARCHAR(255) PRIMARY KEY,
    created_at DATETIME,
    updated_at DATETIME,
    token_hash VARCHAR(255) UNIQUE,
    expires_at DATETIME
);
```

### Password Reset Tokens
```sql
CREATE TABLE password_reset_tokens (
    id VARCHAR(255) PRIMARY KEY,
    created_at DATETIME,
    updated_at DATETIME,
    token VARCHAR(255) UNIQUE,
    user_id VARCHAR(255),
    expires_at DATETIME,
    is_used BOOLEAN DEFAULT FALSE
);
```

## Performance Considerations

- **Redis**: Better performance for high-frequency operations, automatic cleanup
- **Database**: Good performance for moderate usage, requires periodic cleanup of expired tokens

## Migration Guide

To switch from Redis to Database implementations:

1. Set the appropriate environment variables:
   ```bash
   export USE_DATABASE_JWT=true
   export USE_DATABASE_PWRESET=true
   ```

2. Remove Redis dependency from your deployment

3. The application will automatically create the necessary database tables

4. Consider adding a cleanup job for expired tokens:
   ```go
   // Example cleanup function
   func cleanupExpiredTokens(db *gorm.DB) {
       // Clean up expired JWT tokens
       db.Where("expires_at <= ?", time.Now()).Delete(&jwt.BlacklistedToken{})

       // Clean up expired password reset tokens
       db.Where("expires_at <= ?", time.Now()).Delete(&pwreset.PasswordResetToken{})
   }
   ```
