package logger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

type Logger struct {
	*slog.Logger
}

type LoggerConfig struct {
	Level   string
	Service string
}

func NewLogger(config LoggerConfig) *Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler).With("service", config.Service)

	return &Logger{logger}
}

func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return &Logger{l.Logger.With(args...)}
}

func (l *Logger) WithRequestID(ctx context.Context) *Logger {
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		return l.WithAttrs(slog.String("request_id", lc.AwsRequestID))
	}
	return l
}

func HashPhoneNumber(phoneNumber string, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + phoneNumber))
	return hex.EncodeToString(sum[:8])
}
