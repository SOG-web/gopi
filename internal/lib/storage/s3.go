package storage

import (
    "context"
    "fmt"
    "io"
    "net/url"
    "strings"

    minio "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
    client         *minio.Client
    bucket         string
    region         string
    publicBaseURL  string // optional CDN or custom domain; if set, used for public URLs
}

func NewS3Storage(cfg Config) (*S3Storage, error) {
    // Endpoint may be empty for AWS; minio requires an endpoint. For AWS, you can use s3.<region>.amazonaws.com
    endpoint := cfg.S3Endpoint
    if endpoint == "" {
        endpoint = fmt.Sprintf("s3.%s.amazonaws.com", cfg.S3Region)
    }

    cli, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
        Secure: cfg.S3UseSSL,
        Region: cfg.S3Region,
        BucketLookup: func() minio.BucketLookupType {
            if cfg.S3ForcePathStyle {
                return minio.BucketLookupPath
            }
            return minio.BucketLookupAuto
        }(),
    })
    if err != nil {
        return nil, err
    }

    return &S3Storage{
        client:        cli,
        bucket:        cfg.S3Bucket,
        region:        cfg.S3Region,
        publicBaseURL: strings.TrimRight(cfg.S3PublicBaseURL, "/"),
    }, nil
}

func (s *S3Storage) Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
    opts := minio.PutObjectOptions{ContentType: contentType}
    _, err := s.client.PutObject(ctx, s.bucket, key, r, size, opts)
    if err != nil {
        return "", err
    }

    // Build public URL
    if s.publicBaseURL != "" {
        u, _ := url.Parse(s.publicBaseURL)
        // ensure path join without duplicate slashes
        joined := strings.TrimRight(u.String(), "/") + "/" + strings.TrimLeft(key, "/")
        return joined, nil
    }

    // Default to virtual-hosted–style URL on AWS
    // https://<bucket>.s3.<region>.amazonaws.com/<key>
    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
    return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
