package logger

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Logger struct {
	*slog.Logger
}

func SetupLogger(env string) *Logger {
	var handler slog.Handler

	switch env {
	case envLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envDev:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return &Logger{slog.New(handler)}
}

func (l *Logger) Err(err error) slog.Attr {
	return slog.String("error", err.Error())
}
