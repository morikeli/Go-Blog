package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/db"
	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/repositories"
	"github.com/morikeli/golangrestapi/internal/routes"
	"github.com/morikeli/golangrestapi/internal/services"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server   *http.Server
	database *pgxpool.Pool
	redisClient *redis.Client
}

func (a *App) New(cfg *config.Config) (*App, error) {
	// connect to db
	database := db.ConnectDb(cfg.DatabaseURL)

	// connect to redis
	redisClient := db.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)

	// Initialize Cloudinary client once during application startup
	cld, err := cloudinary.NewFromURL(cfg.CloudinaryURL)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary client: %v", err)
	}

	queries := store.New(database)

	// initialize token maker
	tokenMaker := utils.NewTokenMaker(cfg.SecretKey, cfg.JwtIssuer)

	// repository layer
	userRepo := repositories.UserRepository(queries)

	// service layer
	authService := services.NewAuthService(userRepo, tokenMaker)
	userService := services.NewUserService(userRepo)

	// create new handler
	handler := handlers.NewHandler(tokenMaker,
		redisClient,
		cfg,
		cld,
		authService,
		userService,
	)

	// set up http server
	mux := http.NewServeMux()

	// routers
	routes.SetupHealthRoute(mux, handler)
	routes.SetupAuthRoutes(mux, handler)
	routes.SetupUserRoutes(mux, handler)

	// server address
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second, // Timeout for reading request headers & body
		WriteTimeout: 10 * time.Second, // Timeout for writing response
		IdleTimeout:  time.Minute,      // Timeout for keep-alive connections
	}

	return &App{
		server:   server,
		database: database,
		redisClient: redisClient,
	}, nil
}

