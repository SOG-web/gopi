package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"sync"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	host     string
	port     int
	username string
	password string
	from     string

	// Email queue for async processing
	emailQueue chan EmailRequest
	wg         *sync.WaitGroup
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type EmailRequest struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

type OTPEmailData struct {
	Name string
	OTP  string
}

// EmailServiceInterface defines the interface for email operations
type EmailServiceInterface interface {
	SendOTPEmail(email, firstName, otp string) error
	SendWelcomeEmail(email, firstName string) error
	SendPasswordResetEmail(email, resetLink string) error
	SendApologyEmail(email, username string) error
	SendBulkEmail(emails []string, subject, htmlContent string) error
	TestEmailConnection() error
	GetQueueLength() int
}

func NewEmailService(config EmailConfig) *EmailService {
	service := &EmailService{
		host:       config.Host,
		port:       config.Port,
		username:   config.Username,
		password:   config.Password,
		from:       config.From,
		emailQueue: make(chan EmailRequest, 100), // Buffer of 100 emails
		wg:         &sync.WaitGroup{},
	}

	// Start the email worker goroutine (equivalent to Django's EmailThread)
	go service.emailWorker()

	return service
}

// emailWorker processes emails asynchronously (Django's EmailThread equivalent)
func (e *EmailService) emailWorker() {
	for emailReq := range e.emailQueue {
		e.wg.Add(1)
		go func(req EmailRequest) {
			defer e.wg.Done()

			// Create email message
			m := gomail.NewMessage()
			m.SetHeader("From", e.from)
			m.SetHeader("To", req.To...)
			m.SetHeader("Subject", req.Subject)

			if req.IsHTML {
				m.SetBody("text/html", req.Body)
			} else {
				m.SetBody("text/plain", req.Body)
			}

			// Create dialer and send
			d := gomail.NewDialer(e.host, e.port, e.username, e.password)

			if err := d.DialAndSend(m); err != nil {
				// Log error but don't fail the application
				log.Printf("Failed to send email to %v: %v", req.To, err)
			}
		}(emailReq)
	}
}

// SendOTPEmail sends OTP verification email asynchronously (Django's send_register_otp equivalent)
func (e *EmailService) SendOTPEmail(email, firstName, otp string) error {
	// Create email data
	data := OTPEmailData{
		Name: firstName,
		OTP:  otp,
	}

	// Generate HTML content from template
	htmlContent, err := e.generateOTPHTML(data)
	if err != nil {
		return err
	}

	// Queue email for async sending (Django's EmailThread equivalent)
	emailReq := EmailRequest{
		To:      []string{email},
		Subject: "GoPadi Account Verification",
		Body:    htmlContent,
		IsHTML:  true,
	}

	select {
	case e.emailQueue <- emailReq:
		return nil
	default:
		// Queue is full, send synchronously as fallback
		return e.sendEmailSync(emailReq)
	}
}

// SendPasswordResetEmail sends password reset email asynchronously
func (e *EmailService) SendPasswordResetEmail(email, resetLink string) error {
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background-color: #dc3545; color: white; padding: 20px; text-align: center;">
				<h1>Password Reset - GoPadi</h1>
			</div>
			<div style="padding: 20px;">
				<h2>Password Reset Request</h2>
				<p>We received a request to reset your password. Click the link below to reset your password:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px;">Reset Password</a>
				</div>
				<p>This link will expire in 1 hour for security reasons.</p>
				<p>If you didn't request this password reset, please ignore this email.</p>
			</div>
			<div style="background-color: #f8f9fa; padding: 20px; text-align: center; color: #6c757d;">
				<p>This is an automated message, please do not reply to this email.</p>
			</div>
		</body>
		</html>
	`, resetLink)

	// Queue email for async sending
	emailReq := EmailRequest{
		To:      []string{email},
		Subject: "Password Reset - GoPadi",
		Body:    htmlContent,
		IsHTML:  true,
	}

	select {
	case e.emailQueue <- emailReq:
		return nil
	default:
		return e.sendEmailSync(emailReq)
	}
}

// SendWelcomeEmail sends welcome email after verification asynchronously
func (e *EmailService) SendWelcomeEmail(email, firstName string) error {
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background-color: #28a745; color: white; padding: 20px; text-align: center;">
				<h1>Welcome to GoPadi!</h1>
			</div>
			<div style="padding: 20px;">
				<h2>Hello %s!</h2>
				<p>Your account has been successfully verified. Welcome to the GoPadi community!</p>
				<p>You can now:</p>
				<ul>
					<li>Join campaigns and challenges</li>
					<li>Track your fitness activities</li>
					<li>Connect with other fitness enthusiasts</li>
					<li>Make a positive impact through fitness</li>
				</ul>
				<p>Get started by exploring our latest campaigns and challenges.</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="#" style="background-color: #28a745; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px;">Explore Campaigns</a>
				</div>
			</div>
			<div style="background-color: #f8f9fa; padding: 20px; text-align: center; color: #6c757d;">
				<p>Thank you for joining GoPadi!</p>
				<p>&copy; 2024 GoPadi. All rights reserved.</p>
			</div>
		</body>
		</html>
	`, firstName)

	// Queue email for async sending
	emailReq := EmailRequest{
		To:      []string{email},
		Subject: "Welcome to GoPadi!",
		Body:    htmlContent,
		IsHTML:  true,
	}

	select {
	case e.emailQueue <- emailReq:
		return nil
	default:
		return e.sendEmailSync(emailReq)
	}
}

// SendApologyEmail sends apology email (Django equivalent)
func (e *EmailService) SendApologyEmail(email, username string) error {
	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<div style="background-color: #ffc107; color: #212529; padding: 20px; text-align: center;">
				<h1>GoPadi Team - Important Update</h1>
			</div>
			<div style="padding: 20px;">
				<h2>Hello %s,</h2>
				<p>We hope this message finds you well.</p>
				<p>We wanted to reach out to you regarding your experience with GoPadi. Our team is committed to providing the best possible service, and we value your feedback.</p>
				<p>If you've experienced any issues or have suggestions for improvement, please don't hesitate to reach out to us.</p>
				<p>Thank you for being part of the GoPadi community.</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="#" style="background-color: #ffc107; color: #212529; padding: 12px 24px; text-decoration: none; border-radius: 5px;">Contact Support</a>
				</div>
			</div>
			<div style="background-color: #f8f9fa; padding: 20px; text-align: center; color: #6c757d;">
				<p>Best regards,<br>The GoPadi Team</p>
			</div>
		</body>
		</html>
	`, username)

	// Queue email for async sending
	emailReq := EmailRequest{
		To:      []string{email},
		Subject: "GoPadi Team",
		Body:    htmlContent,
		IsHTML:  true,
	}

	select {
	case e.emailQueue <- emailReq:
		return nil
	default:
		return e.sendEmailSync(emailReq)
	}
}

// sendEmailSync sends email synchronously as fallback
func (e *EmailService) sendEmailSync(req EmailRequest) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", req.To...)
	m.SetHeader("Subject", req.Subject)

	if req.IsHTML {
		m.SetBody("text/html", req.Body)
	} else {
		m.SetBody("text/plain", req.Body)
	}

	d := gomail.NewDialer(e.host, e.port, e.username, e.password)
	return d.DialAndSend(m)
}

// generateOTPHTML creates the HTML template for OTP email (matches Django template)
func (e *EmailService) generateOTPHTML(data OTPEmailData) (string, error) {
	tmpl := `
	<html>
	<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
		<div style="background-color: #007bff; color: white; padding: 20px; text-align: center;">
			<h1>GoPadi Account Verification</h1>
		</div>
		<div style="padding: 20px;">
			<h2>Hello {{.Name}}!</h2>
			<p>Thank you for registering with GoPadi. To complete your registration, please use the verification code below:</p>
			
			<div style="background-color: #f8f9fa; padding: 20px; text-align: center; margin: 20px 0; border-radius: 5px;">
				<h1 style="color: #007bff; font-size: 36px; margin: 0; letter-spacing: 5px;">{{.OTP}}</h1>
			</div>
			
			<p>This verification code will expire in 10 minutes for security reasons.</p>
			<p>If you didn't request this verification, please ignore this email.</p>
			
			<p>Welcome to the GoPadi community!</p>
		</div>
		<div style="background-color: #f8f9fa; padding: 20px; text-align: center; color: #6c757d;">
			<p>This is an automated message, please do not reply to this email.</p>
			<p>&copy; 2024 GoPadi. All rights reserved.</p>
		</div>
	</body>
	</html>
	`

	t, err := template.New("otp").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// TestEmailConnection tests the email service configuration
func (e *EmailService) TestEmailConnection() error {
	d := gomail.NewDialer(e.host, e.port, e.username, e.password)

	closer, err := d.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to email server: %v", err)
	}
	defer closer.Close()

	return nil
}

// SendBulkEmail sends email to multiple recipients asynchronously (for announcements)
func (e *EmailService) SendBulkEmail(emails []string, subject, htmlContent string) error {
	// Queue emails for async sending
	for _, email := range emails {
		emailReq := EmailRequest{
			To:      []string{email},
			Subject: subject,
			Body:    htmlContent,
			IsHTML:  true,
		}

		select {
		case e.emailQueue <- emailReq:
			// Successfully queued
		default:
			// Queue is full, send synchronously
			if err := e.sendEmailSync(emailReq); err != nil {
				log.Printf("Failed to send bulk email to %s: %v", email, err)
			}
		}
	}

	return nil
}

// Close gracefully shuts down the email service
func (e *EmailService) Close() {
	close(e.emailQueue)
	e.wg.Wait()
}

// GetQueueLength returns the current number of emails in the queue
func (e *EmailService) GetQueueLength() int {
	return len(e.emailQueue)
}

// NewEmailServiceFactory creates email service based on environment configuration
func NewEmailServiceFactory(config EmailConfig, useLocal bool, logFilePath string) (EmailServiceInterface, error) {
	if useLocal {
		// Use local email service for development - logs to file instead of sending emails
		return NewLocalEmailService(logFilePath)
	}

	// Use production email service
	return NewEmailService(config), nil
}
