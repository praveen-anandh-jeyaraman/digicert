package app

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadConfigFromEnv() (*Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		DatabaseURL: dsn,
		Port:        port,
	}, nil
}
