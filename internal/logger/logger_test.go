package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestMultiHandlerAndServiceAttribute(t *testing.T) {
	var buf bytes.Buffer
	h1 := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})

	// Create MultiHandler
	mh := NewMultiHandler(h1)

	// Create logger with service attribute
	logger := slog.New(mh).With("service", "test-service")

	logger.InfoContext(context.Background(), "hello", "key", "val")

	var logMap map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logMap); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logMap["msg"] != "hello" {
		t.Errorf("Expected msg to be 'hello', got '%v'", logMap["msg"])
	}
	if logMap["level"] != "INFO" {
		t.Errorf("Expected level to be 'INFO', got '%v'", logMap["level"])
	}
	if logMap["service"] != "test-service" {
		t.Errorf("Expected service to be 'test-service', got '%v'", logMap["service"])
	}
	if logMap["key"] != "val" {
		t.Errorf("Expected key to be 'val', got '%v'", logMap["key"])
	}
}
