package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func New() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var out io.Writer = os.Stdout
	if path := os.Getenv("LOG_FILE"); path != "" {
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			out = io.MultiWriter(os.Stdout, f)
		}
	}

	h := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})
	return slog.New(h).With(slog.String("service", "running-tracker"))
}
