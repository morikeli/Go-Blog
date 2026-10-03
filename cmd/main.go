package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Channel to signal server startup errors
	serverErrors := make(chan error, 1)

	// Start server in a non-blocking goroutine
	go func() {
		log.Printf("Server starting on port %s...\n", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 8. Listen for OS signals for graceful shutdown
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal or server startup error is received
	select {
	case err := <-serverErrors:
		log.Fatalf("Server startup failed: %v", err)

	case sig := <-shutdownSig:
		log.Printf("Received signal '%v'. Initiating graceful shutdown...", sig)

		// Create a timeout context for the shutdown process (e.g., 10 seconds)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Stop accepting new connections and wait for active requests to drain
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server force shutdown error: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("Error closing HTTP server: %v", err)
			}
		}

		// Close database and Redis connections AFTER HTTP server has drained requests
		database.Close()
		log.Println("Database connection closed successfully.")

		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		} else {
			log.Println("Redis client connection closed successfully.")
		}

		log.Println("Graceful shutdown complete.")
	}
}
