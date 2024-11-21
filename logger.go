package main

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey int

const (
	ctxLogger ctxKey = iota
)

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger := ctx.Value(ctxLogger); logger != nil {
		return logger.(*slog.Logger)
	}
	return initLogger()
}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}
