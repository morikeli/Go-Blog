package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDb(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(connCtx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(connCtx); err != nil {
		pool.Close()	// close the pool before returning an error
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
