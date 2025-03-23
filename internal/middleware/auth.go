package middleware

import (
	"log"
	"net/http"
	"os"
)

const (
	HeaderAPIKey = "X-API-Key"
)

func RequireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	apiKey := os.Getenv("API_KEY")
	log.Print(apiKey)
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is not set")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next(w, r)
			return
		}
		key := r.Header.Get(HeaderAPIKey)
		if key == "" {
			http.Error(w, "API key is missing", http.StatusUnauthorized)
			return
		}
		if key != apiKey {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
