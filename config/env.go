package config

import (
	"fmt"
	"os"

	"github.com/lpernett/godotenv"
)

type Config struct {
	PublicHost string
	Port       string

	DBUser     string
	DBPassword string
	DBName     string
	DBAddress  string
	DBDriver   string

	LogLevel  string
	LogFile   string
	LogToFile bool
	RunMode   string

	SessionSecret string
	SessionName   string
	SessionSecure bool
	SessionDomain string
	SessionMaxAge int

	// JWT Configuration
	JWTSecret      string
	UseDatabaseJWT bool

	// Email Configuration
	EmailHost     string
	EmailPort     int
	EmailUsername string
	EmailPassword string
	EmailFrom     string
	UseLocalEmail bool
	EmailLogPath  string

	// Redis Configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Password Reset Configuration
	UseDatabasePWReset bool

	// Storage Configuration
	StorageBackend      string // local or s3
	UploadBaseDir       string // e.g. ./uploads
	UploadPublicBaseURL string // e.g. /uploads

	// S3 Configuration
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3UseSSL          bool
	S3ForcePathStyle  bool
	S3PublicBaseURL   string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "8080"),
		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "ecomerce"),
		DBAddress:  getEnv("DB_ADDRESS", fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306"))),

		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFile:   getEnv("LOG_FILE", "logs/app.log"),
		LogToFile: getEnvBool("LOG_FILE_ENABLED", false),
		RunMode:   getEnv("GIN_MODE", "debug"),

		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-change-me"),
		SessionName:   getEnv("SESSION_NAME", "hor_session"),
		SessionSecure: getEnvBool("SESSION_SECURE", false),
		SessionDomain: getEnv("SESSION_DOMAIN", ""),
		SessionMaxAge: getEnvInt("SESSION_MAX_AGE", 86400),

		// JWT Configuration
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret-change-me-in-production"),
		UseDatabaseJWT: getEnvBool("USE_DATABASE_JWT", false),

		// Email Configuration
		EmailHost:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
		EmailPort:     getEnvInt("EMAIL_PORT", 587),
		EmailUsername: getEnv("EMAIL_USERNAME", ""),
		EmailPassword: getEnv("EMAIL_PASSWORD", ""),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@gopadi.com"),
		UseLocalEmail: getEnvBool("USE_LOCAL_EMAIL", true),
		EmailLogPath:  getEnv("EMAIL_LOG_PATH", "./logs/emails.log"),

		// Redis Configuration
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Password Reset Configuration
		UseDatabasePWReset: getEnvBool("USE_DATABASE_PWRESET", false),

		// Storage Configuration
		StorageBackend:      getEnv("STORAGE_BACKEND", "local"),
		UploadBaseDir:       getEnv("UPLOAD_BASE_DIR", "./uploads"),
		UploadPublicBaseURL: getEnv("UPLOAD_PUBLIC_BASE_URL", "/uploads"),

		// S3 Configuration
		S3Endpoint:        getEnv("S3_ENDPOINT", ""),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3Bucket:          getEnv("S3_BUCKET", ""),
		S3AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3UseSSL:          getEnvBool("S3_USE_SSL", true),
		S3ForcePathStyle:  getEnvBool("S3_FORCE_PATH_STYLE", false),
		S3PublicBaseURL:   getEnv("S3_PUBLIC_BASE_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		switch value {
		case "1", "true", "TRUE", "True", "yes", "on", "Y", "y":
			return true
		case "0", "false", "FALSE", "False", "no", "off", "N", "n":
			return false
		default:
			return fallback
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		var n int
		_, err := fmt.Sscan(value, &n)
		if err == nil {
			return n
		}
	}
	return fallback
}
