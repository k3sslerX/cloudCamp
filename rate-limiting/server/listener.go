package server

import (
	"context"
	"net/http"
	"rate-limiting/tokens"
	"sync"
	"time"
)

type contextKey string

const apiKeyContextKey contextKey = "api_key"

var (
	serverInstance *http.Server
	rlInstance     *tokens.RateLimiter
	storage        tokens.DBStorage
	once           sync.Once
)

func init() {
	once.Do(func() {
		rlInstance = tokens.NewRateLimiter()
	})
}

func GetRateLimiter() *tokens.RateLimiter {
	return rlInstance
}

func StartServer() error {
	mux := http.NewServeMux()
	serverInstance = &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		IdleTimeout:  15 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	mux.Handle("/", apiKeyMiddleware(http.HandlerFunc(handleConnection)))

	rl := GetRateLimiter()
	rl.Start()

	return serverInstance.ListenAndServe()
}

func apiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("X-API-KEY")
		if authHeader == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		apiKey := authHeader

		valid, err := storage.ValidateAPIKey(ctx, apiKey)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, apiKeyContextKey, apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Shutdown(ctx context.Context) error {
	if rl := GetRateLimiter(); rl != nil {
		rl.Stop()
	}
	if serverInstance != nil {
		return serverInstance.Shutdown(ctx)
	}
	return nil
}
