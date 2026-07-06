package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/api"
	"github.com/Sush1sui/FNS_BOT/internal/bot"
	"github.com/Sush1sui/FNS_BOT/internal/bot/tickets"
	"github.com/Sush1sui/FNS_BOT/internal/config"
	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// 1. Connect to PostgreSQL
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	poolConfig.MaxConns = int32(getEnvInt("PG_MAX_CONNS", 4))
	poolConfig.MinConns = int32(getEnvInt("PG_MIN_CONNS", 0))
	poolConfig.MaxConnIdleTime = getEnvDuration("PG_MAX_CONN_IDLE", 5*time.Minute)
	poolConfig.MaxConnLifetime = getEnvDuration("PG_MAX_CONN_LIFETIME", 30*time.Minute)
	poolConfig.HealthCheckPeriod = getEnvDuration("PG_HEALTHCHECK_PERIOD", 1*time.Minute)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	keepaliveCtx, keepaliveCancel := context.WithCancel(context.Background())
	defer keepaliveCancel()
	db.StartKeepalive(keepaliveCtx, pool, getEnvDuration("PG_KEEPALIVE_INTERVAL", 5*time.Minute))

	queries := db.New(pool)
	tickets.SetQueries(queries)

	// 2. Connect to Azure Blob Storage
	azureClient := storage.NewAzureClient()
	tickets.SetStorage(azureClient)

	// 3. Mount Router
	mux := api.NewRouter(queries, azureClient, cfg)

	// 4. Start Bot
	bot.StartBot()

	// 5. Start Server
	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	go func() {
		fmt.Printf("Finesse API on port %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	bot.StopBot()
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
