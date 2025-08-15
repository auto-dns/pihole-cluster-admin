package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

func SetupLogger(cfg *config.LoggingConfig) zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	levelStr := strings.ToLower(cfg.Level)
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Str("service", "pihole_cluster_admin").
		Logger()

	return logger
}

type modeKeyT struct{}

var modeKey modeKeyT

type Mode int

const (
	ModeDefault Mode = iota
	ModeTrace
	ModeDebug
)

func WithMode(ctx context.Context, m Mode) context.Context {
	return context.WithValue(ctx, modeKey, m)
}

func ModeFrom(ctx context.Context) Mode {
	if v, ok := ctx.Value(modeKey).(Mode); ok {
		return v
	}
	return ModeDefault
}

func WithContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

func From(ctx context.Context, fallback zerolog.Logger) zerolog.Logger {
	if l := zerolog.Ctx(ctx); l != nil {
		return *l
	}
	return fallback
}

func Event(ctx context.Context, fallback zerolog.Logger) *zerolog.Event {
	l := From(ctx, fallback)
	switch ModeFrom(ctx) {
	case ModeTrace:
		return l.Trace()
	case ModeDebug, ModeDefault:
		return l.Debug()
	default:
		return l.Debug()
	}
}
