package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "super-secret-test-key-12345")

	called := false
	handler := RequireAPIKey(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows /health without API key", func(t *testing.T) {
		called = false
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)

		if rr.Code != http.StatusOK || !called {
			t.Fatalf("expected 200 and handler called, got %d (called=%v)", rr.Code, called)
		}
	})

	t.Run("rejects missing API key", func(t *testing.T) {
		called = false
		req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
		rr := httptest.NewRecorder()
		handler(rr, req)

		if rr.Code != http.StatusUnauthorized || called {
			t.Fatalf("expected 401 and handler not called, got %d (called=%v)", rr.Code, called)
		}
	})

	t.Run("rejects wrong API key", func(t *testing.T) {
		called = false
		req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
		req.Header.Set(HeaderAPIKey, "wrong-key")
		rr := httptest.NewRecorder()
		handler(rr, req)

		if rr.Code != http.StatusUnauthorized || called {
			t.Fatalf("expected 401 and handler not called, got %d (called=%v)", rr.Code, called)
		}
	})

	t.Run("accepts valid API key", func(t *testing.T) {
		called = false
		req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
		req.Header.Set(HeaderAPIKey, "super-secret-test-key-12345")
		rr := httptest.NewRecorder()
		handler(rr, req)

		if rr.Code != http.StatusOK || !called {
			t.Fatalf("expected 200 and handler called, got %d (called=%v)", rr.Code, called)
		}
	})
}
