package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/morikeli/golangrestapi/internal/config"
	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/routes"
)

func main() {
	// load project config
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load project configuration file: %v", err)
	}

	// set up http server
	mux := http.NewServeMux()

	// create new handler
	handler := handlers.NewHandler()
	
	// routers
	routes.SetupHealthRoute(mux, handler)

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