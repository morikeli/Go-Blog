package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DatabaseURL string
	Environment string
	LogLevel string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("Error getting environment variables! \n %v", err)
	}

	config := &Config {
		ServerPort: os.Getenv("SERVER_PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Environment: os.Getenv("ENVIRONMENT"),	// dev, prod environment
		LogLevel: os.Getenv("LOG_LEVEL"),	// debug, info, warn, error
	}

	return config, nil
	
}