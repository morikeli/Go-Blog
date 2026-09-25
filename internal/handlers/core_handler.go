package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	Db         *pgxpool.Pool
	Queries    *store.Queries
	TokenMaker *utils.TokenMaker
	Redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, query *store.Queries, tokenMaker *utils.TokenMaker, redis *redis.Client) *Handler {
	return &Handler{
		Db:         db,
		Queries:    query,
		TokenMaker: tokenMaker,
		Redis:      redis,
	}
}
