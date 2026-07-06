package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func StartKeepalive(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	if pool == nil || interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		pingPool(pool)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pingPool(pool)
			}
		}
	}()
}

func pingPool(pool *pgxpool.Pool) {
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Printf("db keepalive ping failed: %v", err)
	}
}
