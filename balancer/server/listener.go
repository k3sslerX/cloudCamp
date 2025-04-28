package server

import (
	"balancer/backends"
	"balancer/redirections"
	"net/http"
	"time"
)

func StartServer() error {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		IdleTimeout:  15 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	mux.Handle("/", http.HandlerFunc(redirections.HandleConnection))

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	go backends.HealthChecker()

	return nil
}
