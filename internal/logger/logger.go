package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"gopi.com/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a slog.Logger that logs to both terminal (human-readable) and file (JSON) with rotation.
func New(cfg config.Config) *slog.Logger {
	// Terminal handler (colorized)
	termHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      parseLevel(cfg.LogLevel),
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})

	// Ensure log directory exists
	if dir := filepath.Dir(cfg.LogFile); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	// Optionally add file handler (JSON) with rotation
	handlers := []slog.Handler{termHandler}
	if cfg.LogToFile {
		lj := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    25, // megabytes
			MaxBackups: 5,
			MaxAge:     28,   // days
			Compress:   true, // gzip rotated files
		}
		fileWriter := io.Writer(lj)
		fileHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)})
		handlers = append(handlers, fileHandler)
	}

	// Fan-out handler: send to terminal and optionally file
	multi := NewMultiHandler(handlers...)
	return slog.New(multi)
}

// parseLevel returns the slog.Level based on a string value.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// MultiHandler fans out records to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(_ context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(context.Background(), level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		// Clone record time and attrs for each handler
		rec := slog.Record{
			Time:    time.Now(),
			Message: r.Message,
			Level:   r.Level,
		}
		r.Attrs(func(a slog.Attr) bool { rec.AddAttrs(a); return true })
		if err := h.Handle(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
