package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	fmt.Println("Successfully connected to Redis!")
	return rdb, nil
}
