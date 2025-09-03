# Local Development Email Service

This directory contains a local development email service that logs emails to a file instead of sending them. This is perfect for frontend development and testing.

## Files

- `email_service.go` - Production email service using SMTP
- `local_email_service.go` - Local development service that logs to file

## Usage

The email service automatically switches between local and production modes based on the `USE_LOCAL_EMAIL` environment variable.

### Environment Configuration

**For Local Development (Email Logging):**

```bash
USE_LOCAL_EMAIL=true
```

**For Production (Real Email Sending):**

```bash
USE_LOCAL_EMAIL=false
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@gopadi.com
```

### Automatic Service Selection

The system automatically chooses the appropriate email service:

- **USE_LOCAL_EMAIL=true**: Uses `LocalEmailService` that logs to `./logs/emails.log`
- **USE_LOCAL_EMAIL=false**: Uses `EmailService` that sends real emails via SMTP

This follows the same factory pattern as the password reset service (`USE_DATABASE_PWRESET`).

### Log File Location

The local email service will create a log file at the specified path. Make sure the directory exists or it will be created automatically.

Example log locations:

- `./logs/emails.log`
- `/tmp/gopi_emails.log`
- `./email_logs/dev.log`

## What Gets Logged

The local service logs all email operations with detailed information:

### OTP Emails

```
=========================================
OTP EMAIL REQUEST
=========================================
To: user@example.com
Name: John Doe
OTP CODE: 123456
Timestamp: 2024-01-15T10:30:45Z
=========================================
COPY THIS OTP CODE FOR TESTING:
OTP: 123456
=========================================
```

### Password Reset Emails

```
=========================================
PASSWORD RESET EMAIL REQUEST
=========================================
To: user@example.com
Reset Link: https://app.com/reset?token=abc123
Timestamp: 2024-01-15T10:30:45Z
=========================================
COPY THIS RESET LINK FOR TESTING:
LINK: https://app.com/reset?token=abc123
=========================================
```

### Other Email Types

- Welcome emails
- Apology emails
- Bulk emails (with recipient count and content preview)

## Benefits for Frontend Development

1. **No Email Dependencies**: No need to configure SMTP servers
2. **Easy Testing**: OTP codes and reset links are clearly logged
3. **Offline Development**: Works without internet connection
4. **No Spam**: Doesn't send actual emails during testing
5. **Audit Trail**: All email attempts are logged for debugging

## Testing Tips

1. **Monitor the log file** during development:

   ```bash
   tail -f ./logs/emails.log
   ```

2. **Extract OTP codes** for testing:

   ```bash
   grep "OTP:" ./logs/emails.log | tail -1
   ```

3. **Extract reset links** for testing:
   ```bash
   grep "LINK:" ./logs/emails.log | tail -1
   ```

## Production Deployment

Remember to switch back to the production email service for staging/production environments. The local service should never be used in production as it doesn't actually send emails.
