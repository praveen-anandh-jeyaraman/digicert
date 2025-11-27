package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/app"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/handler"
    // "github.com/praveen-anandh-jeyaraman/digicert/internal/logger"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/repo"
    "github.com/praveen-anandh-jeyaraman/digicert/internal/service"
    _ "github.com/praveen-anandh-jeyaraman/digicert/docs"
)

// @title           DigiCert Book API
// @version         1.0
// @description     A RESTful API for managing books and borrowing system
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
    ctx := context.Background()

    cfg, err := app.LoadConfigFromEnv()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Initialize CloudWatch logger
    // if err := logger.Initialize(cfg.CloudWatchLogGroup, cfg.CloudWatchLogStream, cfg.EnableCloudWatch); err != nil {
    //     log.Printf("Warning: CloudWatch initialization failed: %v", err)
    // }
    // defer logger.GetLogger().Close()
    // log.Printf("Logger initialized - CloudWatch: %v", cfg.EnableCloudWatch)

    stdLogger := app.NewStdLogger()

    dbpool, err := app.NewDBPool(ctx, cfg)
    if err != nil {
        stdLogger.Fatalf("db connect failed: %v", err)
    }
    defer dbpool.Close()

    // Initialize repositories
    bookRepo := repo.NewBookRepo(dbpool)
    userRepo := repo.NewUserRepo(dbpool)
    bookingRepo := repo.NewBookingRepo(dbpool)

    // Initialize services
    bookSvc := service.NewBookService(bookRepo)
    userSvc := service.NewUserService(userRepo)
    bookingSvc := service.NewBookingService(bookingRepo, bookRepo, userRepo)
    authSvc := service.NewAuthService("your-secret-key-change-this", 24*time.Hour)

    // Initialize handlers
    bookHandler := handler.NewBookHandler(bookSvc)
    userHandler := handler.NewUserHandler(userSvc)
    bookingHandler := handler.NewBookingHandler(bookingSvc)
    authHandler := handler.NewAuthHandler(authSvc, userSvc)

    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(handler.RequestIDMiddleware)
    r.Use(handler.LoggingMiddleware)

    // Health checks (PUBLIC)
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"healthy"}`))
    })

    r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
        if err := dbpool.Ping(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            _, _ = w.Write([]byte(`{"status":"not_ready"}`))
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"status":"ready"}`))
    })

    // Auth endpoints (PUBLIC)
    r.Post("/auth/register", userHandler.Register)
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/refresh", authHandler.Refresh)
    r.Post("/auth/admin-register", userHandler.RegisterAdmin) 

    // User endpoints (PROTECTED - ALL USERS)
    r.Group(func(r chi.Router) {
        r.Use(handler.AuthMiddleware(authSvc))
        r.Get("/users/me", userHandler.GetProfile)
        r.Put("/users/me", userHandler.UpdateProfile)
    })

    // Admin endpoints (PROTECTED - ADMIN ONLY)
    r.Group(func(r chi.Router) {
        r.Use(handler.AuthMiddleware(authSvc))
        r.Use(handler.AdminMiddleware)

        // Book CRUD (admin only)
        r.Route("/admin/books", func(r chi.Router) {
            r.Get("/", bookHandler.List)
            r.Post("/", bookHandler.Create)
            r.Get("/{id}", bookHandler.Get)
            r.Put("/{id}", bookHandler.Update)
            r.Delete("/{id}", bookHandler.Delete)
        })

        // User management (admin only)
        r.Route("/admin/users", func(r chi.Router) {
            r.Get("/", userHandler.ListUsers)
            r.Get("/{id}", userHandler.GetUser)
            r.Delete("/{id}", userHandler.DeleteUser)
        })

        // View all bookings (admin only)
        r.Get("/admin/bookings", bookingHandler.ListAllBookings)
    })

    // Public book viewing
    r.Get("/books", bookHandler.List)

    // User borrowing endpoints (PROTECTED - ALL USERS)
    r.Group(func(r chi.Router) {
        r.Use(handler.AuthMiddleware(authSvc))

        // Book viewing (any user)
        r.Get("/books/{id}", bookHandler.Get)

        // Borrowing (any user)
        r.Route("/bookings", func(r chi.Router) {
            r.Get("/", bookingHandler.GetMyBookings)
            r.Post("/", bookingHandler.Borrow)
            r.Get("/{id}", bookingHandler.GetBooking)
            r.Post("/{id}/return", bookingHandler.Return)
        })
    })
 port := cfg.Port
if port == "" { port = "8080" }
if strings.Contains(port, ":") {
    parts := strings.Split(port, ":")
    port = parts[len(parts)-1]
}
// Force listen on all interfaces
addr := ":" + port

    srv := &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        log.Printf("starting server on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe(): %v", err)
        }
    }()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    log.Println("shutting down")

    ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctxShutdown); err != nil {
        log.Fatalf("server shutdown failed: %v", err)
    }
    log.Println("server stopped")
}