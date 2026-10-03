package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/db"
	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/repositories"
	"github.com/morikeli/golangrestapi/internal/routes"
	"github.com/morikeli/golangrestapi/internal/services"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func main() {
	// load project config
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load project configuration file: %v", err)
	}

	// Initialize Cloudinary client once during application startup
	cld, err := cloudinary.NewFromURL(cfg.CloudinaryURL)
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary client: %v", err)
	}

	// connect to db
	database := db.ConnectDb(cfg.DatabaseURL)
	defer database.Close()

	// connect to redis
	redisClient := db.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)
	defer redisClient.Close()

	queries := store.New(database)

	// create new handler & token maker
	tokenMaker := utils.NewTokenMaker(cfg.SecretKey, cfg.JwtIssuer)

	// repository layer
	userRepo := repositories.UserRepository(queries)

	// service layer
	authService := services.NewAuthService(userRepo, tokenMaker)
	userService := services.NewUserService(userRepo)

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
		Addr:    serverAddr,
		Handler: mux,
		ReadTimeout:  10 * time.Second,	// Timeout for reading request headers & body
		WriteTimeout: 10 * time.Second,	// Timeout for writing response
		IdleTimeout:  time.Minute,	// Timeout for keep-alive connections
	}

	fmt.Printf("Server started on port %s\n", cfg.ServerPort)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start! Error: %v", err)
	}
}
