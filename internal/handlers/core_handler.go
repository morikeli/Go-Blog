package handlers

import (
	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/services"
	"github.com/morikeli/golangrestapi/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	TokenMaker  *utils.TokenMaker
	Redis       *redis.Client
	Config      *config.Config
	AuthService *services.AuthService
}

func NewHandler(
	tokenMaker *utils.TokenMaker,
	redis *redis.Client,
	cfg *config.Config,
	authService *services.AuthService,
) *Handler {
	return &Handler{
		TokenMaker:  tokenMaker,
		Redis:       redis,
		Config:      cfg,
		AuthService: authService,
	}
}
