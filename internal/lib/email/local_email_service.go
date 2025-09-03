package email

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LocalEmailService implements EmailServiceInterface for local development
// Instead of sending emails, it logs them to a file for testing purposes
type LocalEmailService struct {
	logFile *os.File
	logger  *log.Logger
	mu      sync.Mutex
}

// NewLocalEmailService creates a new local email service that logs to file
func NewLocalEmailService(logFilePath string) (*LocalEmailService, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file in append mode
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create logger
	logger := log.New(file, "[EMAIL] ", log.LstdFlags|log.LUTC)

	service := &LocalEmailService{
		logFile: file,
		logger:  logger,
	}

	// Log service initialization
	service.logger.Println("=========================================")
	service.logger.Println("LOCAL EMAIL SERVICE INITIALIZED")
	service.logger.Println("=========================================")
	service.logger.Printf("Log file: %s\n", logFilePath)
	service.logger.Println("All emails will be logged instead of sent")
	service.logger.Println("=========================================")

	return service, nil
}

// SendOTPEmail logs OTP email details instead of sending
func (l *LocalEmailService) SendOTPEmail(email, firstName, otp string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("OTP EMAIL REQUEST")
	l.logger.Println("=========================================")
	l.logger.Printf("To: %s\n", email)
	l.logger.Printf("Name: %s\n", firstName)
	l.logger.Printf("OTP CODE: %s\n", otp)
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")
	l.logger.Println("COPY THIS OTP CODE FOR TESTING:")
	l.logger.Printf("OTP: %s\n", otp)
	l.logger.Println("=========================================")

	return nil
}

// SendPasswordResetEmail logs password reset email details
func (l *LocalEmailService) SendPasswordResetEmail(email, resetLink string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("PASSWORD RESET EMAIL REQUEST")
	l.logger.Println("=========================================")
	l.logger.Printf("To: %s\n", email)
	l.logger.Printf("Reset Link: %s\n", resetLink)
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")
	l.logger.Println("COPY THIS RESET LINK FOR TESTING:")
	l.logger.Printf("LINK: %s\n", resetLink)
	l.logger.Println("=========================================")

	return nil
}

// SendWelcomeEmail logs welcome email details
func (l *LocalEmailService) SendWelcomeEmail(email, firstName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("WELCOME EMAIL REQUEST")
	l.logger.Println("=========================================")
	l.logger.Printf("To: %s\n", email)
	l.logger.Printf("Name: %s\n", firstName)
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")

	return nil
}

// SendApologyEmail logs apology email details
func (l *LocalEmailService) SendApologyEmail(email, username string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("APOLOGY EMAIL REQUEST")
	l.logger.Println("=========================================")
	l.logger.Printf("To: %s\n", email)
	l.logger.Printf("Username: %s\n", username)
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")

	return nil
}

// SendBulkEmail logs bulk email details
func (l *LocalEmailService) SendBulkEmail(emails []string, subject, htmlContent string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("BULK EMAIL REQUEST")
	l.logger.Println("=========================================")
	l.logger.Printf("Subject: %s\n", subject)
	l.logger.Printf("Recipients: %v\n", emails)
	l.logger.Printf("Number of recipients: %d\n", len(emails))
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")
	l.logger.Println("HTML Content Preview:")
	l.logger.Println(htmlContent[:min(500, len(htmlContent))]) // First 500 chars
	if len(htmlContent) > 500 {
		l.logger.Println("... (truncated)")
	}
	l.logger.Println("=========================================")

	return nil
}

// TestEmailConnection always returns success for local service
func (l *LocalEmailService) TestEmailConnection() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("EMAIL CONNECTION TEST")
	l.logger.Println("=========================================")
	l.logger.Println("Local email service - connection test passed")
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")

	return nil
}

// GetQueueLength returns 0 for local service (no actual queue)
func (l *LocalEmailService) GetQueueLength() int {
	return 0
}

// Close closes the log file
func (l *LocalEmailService) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Println("=========================================")
	l.logger.Println("LOCAL EMAIL SERVICE SHUTTING DOWN")
	l.logger.Printf("Timestamp: %s\n", time.Now().UTC().Format(time.RFC3339))
	l.logger.Println("=========================================")

	return l.logFile.Close()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
