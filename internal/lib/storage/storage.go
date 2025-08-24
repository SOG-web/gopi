package storage

import (
    "context"
    "io"
)

// Storage abstracts upload/delete operations and returns a public URL for saved objects.
// Implementations must be safe for concurrent use.
type Storage interface {
    // Save stores content at the provided key (e.g., "profile/filename.jpg") and returns a public URL.
    Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (publicURL string, err error)
    // Delete removes an object at key. Should be idempotent.
    Delete(ctx context.Context, key string) error
}

// Config is a generic storage configuration. Concrete backends may use a subset.
type Config struct {
    // Backend: "local" or "s3"
    Backend string

    // Local settings
    LocalBaseDir          string // filesystem base dir, e.g. ./uploads
    LocalPublicBaseURL    string // public base URL/prefix served by the API, e.g. /uploads

    // S3 settings (also compatible with MinIO)
    S3Endpoint        string // optional; if empty uses AWS SDK endpoint resolution or MinIO endpoint
    S3Region          string
    S3Bucket          string
    S3AccessKeyID     string
    S3SecretAccessKey string
    S3UseSSL          bool
    S3ForcePathStyle  bool
    S3PublicBaseURL   string // optional CDN/CloudFront domain; if set, used to construct public URL
}
