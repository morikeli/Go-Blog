package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morikeli/golangrestapi/internal/store"
)

type Handler struct {
	Db *pgxpool.Pool
	Queries *store.Queries
}

func NewHandler(db *pgxpool.Pool, query *store.Queries) *Handler {
	return &Handler{
		Db:      db,
		Queries: query,
	}
}