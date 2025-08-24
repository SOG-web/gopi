package storage

import (
    "context"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

type LocalStorage struct {
    baseDir       string
    publicBaseURL string // e.g. /uploads
}

func NewLocalStorage(baseDir, publicBaseURL string) *LocalStorage {
    return &LocalStorage{baseDir: baseDir, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
}

func (s *LocalStorage) Save(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
    // Clean and ensure directories exist
    cleanKey := filepath.ToSlash(filepath.Clean(key))
    absPath := filepath.Join(s.baseDir, filepath.FromSlash(cleanKey))

    if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
        return "", fmt.Errorf("prepare dir: %w", err)
    }

    // Create file
    f, err := os.Create(absPath)
    if err != nil {
        return "", fmt.Errorf("create file: %w", err)
    }
    defer f.Close()

    if _, err := io.Copy(f, r); err != nil {
        return "", fmt.Errorf("write file: %w", err)
    }

    // Construct public URL
    public := s.publicBaseURL + "/" + cleanKey
    return public, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
    cleanKey := filepath.ToSlash(filepath.Clean(key))
    absPath := filepath.Join(s.baseDir, filepath.FromSlash(cleanKey))
    if err := os.Remove(absPath); err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return nil
        }
        return err
    }
    return nil
}
