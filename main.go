package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/db"
	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/routes"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func main() {
	// load project config
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load project configuration file: %v", err)
	}

	// connect to db
	database := db.ConnectDb(cfg.DatabaseURL)
	defer database.Close()

	queries := store.New(database)

	// set up http server
	mux := http.NewServeMux()

	// create new handler & token maker
	tokenMaker := utils.NewTokenMaker(cfg.SecretKey, cfg.JwtIssuer)
	handler := handlers.NewHandler(database, queries, tokenMaker)

	// routers
	routes.SetupHealthRoute(mux, handler)
	routes.SetupAuthRoutes(mux, handler)

	// server address
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr: serverAddr,
		Handler: mux,
	}

	fmt.Printf("Server started on port %s\n", cfg.ServerPort)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start! Error: %v", err)
	}
}