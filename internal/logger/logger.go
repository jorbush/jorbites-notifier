package logger

import (
	"context"
	"log/slog"
	"os"

	adapter "github.com/axiomhq/axiom-go/adapters/slog"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
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

func Init() func() {
	var handlers []slog.Handler

	// Add JSON stdout handler
	handlers = append(handlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var axiomHandler *adapter.Handler
	token := os.Getenv("AXIOM_TOKEN")
	dataset := os.Getenv("AXIOM_DATASET")

	if token != "" && dataset != "" {
		var err error
		axiomHandler, err = adapter.New(
			adapter.SetDataset(dataset),
		)
		if err != nil {
			slog.Error("Failed to initialize Axiom handler", "error", err)
		} else {
			handlers = append(handlers, axiomHandler)
		}
	}

	logger := slog.New(&MultiHandler{handlers: handlers})
	slog.SetDefault(logger.With("service", "jorbites-notifier"))

	return func() {
		if axiomHandler != nil {
			axiomHandler.Close()
		}
	}
}
