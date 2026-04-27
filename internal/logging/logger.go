package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

func New(level string) *slog.Logger {
	return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:     levelFor(level),
		AddSource: true,
	}))
}

func levelFor(value string) slog.Leveler {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
