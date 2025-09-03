# Environment Configuration Example

Copy these settings to your `.env` file for local development with email logging.

## For Local Development (Email Logging)

```bash
# Set this to true to use local email service that logs instead of sending
USE_LOCAL_EMAIL=true
```

When `USE_LOCAL_EMAIL=true`, all emails will be logged to `./logs/emails.log` instead of being sent via SMTP.

## For Production (Real Email Sending)

```bash
# Set this to false to use production email service
USE_LOCAL_EMAIL=false

# Configure your SMTP settings
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@gopadi.com
```

## Complete .env Example

```bash
# Database
DB_DRIVER=sqlite
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=ecommerce

# Email Configuration
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@gopadi.com
USE_LOCAL_EMAIL=true  # Set to false for production

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=dev-jwt-secret-change-me-in-production
USE_DATABASE_JWT=false

# Password Reset
USE_DATABASE_PWRESET=false

# Logging
LOG_LEVEL=info
LOG_FILE=logs/app.log
LOG_FILE_ENABLED=false
GIN_MODE=debug

# Server
PUBLIC_HOST=http://localhost
PORT=8080

# Storage
STORAGE_BACKEND=local
UPLOAD_BASE_DIR=./uploads
UPLOAD_PUBLIC_BASE_URL=/uploads
```

## How It Works

- **USE_LOCAL_EMAIL=true**: Emails are logged to `./logs/emails.log`
- **USE_LOCAL_EMAIL=false**: Emails are sent via SMTP using the EMAIL_* settings

The system automatically chooses the appropriate email service based on this environment variable, just like how password reset service switches between Redis and database implementations.
