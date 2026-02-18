package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
)

var ERROR_INVALID_PATH = errors.New("invalid path to the file / file does not exist")

type Handler struct {
	slog.Handler
	l *log.Logger
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	fields := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(fields, "", "	")
	if err != nil {
		return fmt.Errorf("Handle: %s", err.Error())
	}

	timeStr := r.Time.Format("[15:10:05.000]")
	h.l.Println(timeStr, r.Level, r.Message, string(b))
	return nil
}

func NewHandler(out io.Writer, opts *slog.HandlerOptions) *Handler {
	h := &Handler{
		Handler: slog.NewJSONHandler(out, opts),
		l:       log.New(out, "", 0),
	}
	return h
}
