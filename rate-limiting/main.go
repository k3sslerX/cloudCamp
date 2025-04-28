package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rate-limiting/server"
	"rate-limiting/tokens"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error parsing PostgreSQL config:", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Error creating connection pool:", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatal("Could not ping database:", err)
	}

	storage := tokens.NewDBStorage(pool)

	rl := server.GetRateLimiter()
	if err := storage.LoadConfigs(rl); err != nil {
		log.Fatal("Error loading rate limit configs:", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error)
	go func() {
		if len(os.Args) == 2 {
			serverErr <- server.StartServer(storage, os.Args[1])
		} else {
			serverErr <- server.StartServer(storage)
		}
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server error:", err)
		}
	case <-done:
		log.Println("Server is shutting down...")

		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println("Error during server shutdown:", err)
		}

		log.Println("Server stopped")
	}
}
