package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tanapon-gau/miniflow/worker-go/internal/db"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/queue"
	"github.com/Tanapon-gau/miniflow/worker-go/internal/worker"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://miniflow:miniflow@localhost:5433/miniflow")
	redisAddr := env("REDIS_ADDR", "localhost:6380")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := db.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer database.Close()

	q := queue.New(redisAddr)
	defer q.Close()

	log.Println("worker-go starting, consuming tasks:shell and tasks:http")
	worker.New(database, q).Run(ctx)
	log.Println("worker-go stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
