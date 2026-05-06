package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tanapon-gau/miniflow/scheduler/internal/db"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/queue"
	"github.com/Tanapon-gau/miniflow/scheduler/internal/scheduler"
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

	log.Println("scheduler starting, poll interval 5s")
	scheduler.New(database, q, 5*time.Second).Run(ctx)
	log.Println("scheduler stopped")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
