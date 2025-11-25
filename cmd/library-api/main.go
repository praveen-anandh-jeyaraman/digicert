package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/app"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/handler"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
	_ "github.com/praveen-anandh-jeyaraman/digicert/docs"
)
// @title           DigiCert Book API
// @version         1.0
// @description     A RESTful API for managing books
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes http https

func main() {
    ctx := context.Background()

    cfg, err := app.LoadConfigFromEnv()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    logger := app.NewStdLogger()

    dbpool, err := app.NewDBPool(ctx, cfg)
    if err != nil {
        logger.Fatalf("db connect failed: %v", err)
    }
    defer dbpool.Close()

    bookRepo := repo.NewBookRepo(dbpool)
    bookSvc := service.NewBookService(bookRepo)
    bookHandler := handler.NewBookHandler(bookSvc)

    r := chi.NewRouter()
    
    // Add middleware for production-grade observability
    r.Use(handler.RequestIDMiddleware)
    r.Use(handler.RateLimitMiddleware(100))  // 100 req/sec per IP
    r.Use(handler.LoggingMiddleware)
    r.Use(handler.RecoveryMiddleware)
    r.Use(middleware.Compress(5))  // Gzip compression

	r.Get("/swagger/*", httpSwagger.WrapHandler)


    // Health checks
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
 		if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
            logger.Printf("failed to write response: %v", err)
        }
    })

    r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
        // Check database connectivity
        if err := dbpool.Ping(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            _, _ = w.Write([]byte(`{"status":"not_ready"}`))
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ready"}`))
    })

    // Books API
    r.Route("/books", func(r chi.Router) {
        r.Get("/", bookHandler.List)
        r.Post("/", bookHandler.Create)
        r.Get("/{id}", bookHandler.Get)
        r.Put("/{id}", bookHandler.Update)
        r.Delete("/{id}", bookHandler.Delete)
    })

    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        logger.Printf("starting server on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("ListenAndServe(): %v", err)
        }
    }()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    logger.Println("shutting down")
    
    ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctxShutdown); err != nil {
        logger.Fatalf("server shutdown failed: %v", err)
    }
    logger.Println("server stopped")
}