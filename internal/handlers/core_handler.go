package handlers

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/services"
	"github.com/morikeli/golangrestapi/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	TokenMaker  *utils.TokenMaker
	Redis       *redis.Client
	Config      *config.Config
	Cloudinary  *cloudinary.Cloudinary
	AuthService *services.AuthService
	UserService *services.UserService
}

func NewHandler(
	tokenMaker *utils.TokenMaker,
	redis *redis.Client,
	cfg *config.Config,
	cloudinary *cloudinary.Cloudinary,
	authService *services.AuthService,
	userService *services.UserService,
) *Handler {
	return &Handler{
		TokenMaker:  tokenMaker,
		Redis:       redis,
		Config:      cfg,
		Cloudinary:  cloudinary,
		AuthService: authService,
		UserService: userService,
	}
}
