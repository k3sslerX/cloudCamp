package server

import (
	"net/http"
)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	rl := GetRateLimiter()
	if rl == nil {
		http.Error(w, "Rate limiter not initialized", http.StatusInternalServerError)
		return
	}

	authHeader := r.Header.Get("X-API-Key")
	if authHeader == "" {
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	apiKey := authHeader

	allowed, err := rl.Allow(apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	_, err = w.Write([]byte("Request processed successfully"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
