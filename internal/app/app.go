package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App is the central application container.
// It wires together config, db pool, logger and other shared resources.
type App struct {
	Config *Config
	Logger *log.Logger
	DB     *pgxpool.Pool
}

// NewStdLogger returns a simple standard library logger writing to stdout.
func NewStdLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}

// New creates a fully initialized App instance.
// It loads configuration, sets up logging and initializes the DB connection pool.
func New(ctx context.Context) (*App, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := NewStdLogger()

	db, err := NewDBPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}, nil
}

// Close releases resources gracefully.
// main.go should call app.Close() during shutdown.
func (a *App) Close(ctx context.Context) error {
	if a.DB != nil {
		a.DB.Close()
	}
	// add other closers here (redis, grpc clients, tracers, metrics)
	return nil
}
