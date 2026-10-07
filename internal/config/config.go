package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	Environment   string
	LogLevel      string
	SecretKey     string
	JwtIssuer     string
	RedisAddr     string
	RedisPassword string
	CloudinaryURL string
	CORSOrigins   []string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("Error getting environment variables! \n %v", err)
	}

	config := &Config{
		JwtIssuer:     os.Getenv("JWT_ISSUER"),
		SecretKey:     os.Getenv("SECRET_KEY"),
		ServerPort:    os.Getenv("SERVER_PORT"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Environment:   os.Getenv("ENVIRONMENT"), // dev, prod environment
		LogLevel:      os.Getenv("LOG_LEVEL"),   // debug, info, warn, error
		RedisAddr:     os.Getenv("REDIS_ADDRESS"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		CloudinaryURL: os.Getenv("CLOUDINARY_URL"),
		CORSOrigins:   strings.Split(os.Getenv("CORS_ORIGIN"), ","),
	}

	return config, nil

}
