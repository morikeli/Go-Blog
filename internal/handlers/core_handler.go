package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

type Handler struct {
	Db         *pgxpool.Pool
	Queries    *store.Queries
	TokenMaker *utils.TokenMaker
}

func NewHandler(db *pgxpool.Pool, query *store.Queries, tokenMaker *utils.TokenMaker) *Handler {
	return &Handler{
		Db:         db,
		Queries:    query,
		TokenMaker: tokenMaker,
	}
}
