package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartKeepalive pings the pool every 5 minutes to prevent idle shutdown.
func StartKeepalive(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	if pool == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
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
	pool.Ping(pingCtx) //nolint:errcheck
}
