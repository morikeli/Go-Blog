package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDb(databaseURL string) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		log.Fatal("[ERROR]: Failed to connect to database", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("[ERROR]: Failed to ping database", err)
	}

	return pool
}
