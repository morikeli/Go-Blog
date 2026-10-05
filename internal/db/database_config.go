package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDb(databaseURL string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()	// close the pool before returning an error
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
