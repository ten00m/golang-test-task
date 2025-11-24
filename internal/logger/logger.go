package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates a new structured logger using slog with JSON output
func New(output io.Writer) *slog.Logger {
	if output == nil {
		output = os.Stdout
	}

	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// NewDev creates a new logger for development with text output
func NewDev() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
