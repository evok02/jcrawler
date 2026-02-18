package logger

import (
	"log/slog"
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	handler := NewHandler(os.Stdout, nil)

	errLogger := slog.New(handler)

	errLogger.Error("something happend",
		slog.Int("err", 1),
		slog.Group("trace",
			slog.String("error", "capcity is full"),
			slog.String("error", "stack overflow")),
		slog.Group("user",
			slog.String("name", "hlob"),
			slog.String("surname", "ter")))
}
