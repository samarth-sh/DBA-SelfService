package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend/internals/database"
	"go-backend/routes"
	"github.com/rs/cors"
    "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type HealthInfo struct {
    AppName string `json:"app_name"`
    Version string `json:"version"`
    Status  string `json:"status"`
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
    healthInfo := HealthInfo{
        AppName: "DBASelfService-Go Backend",
        Version: "1.0.0",
        Status:  "OK",
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(healthInfo); err != nil {
        log.Error().Err(err).Msg("Failed to encode health info")
    }
}

func main() {

    log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
    log.Info().Msg("Starting server...")

    db, err := database.ConnectPostgres()
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL database")
    }
    defer func() {
        if err := db.Close(); err != nil {
            log.Error().Err(err).Msg("Error closing PostgreSQL database connection")
        }
    }()
    log.Info().Msg("Connected to PostgreSQL database successfully")
    msdb, err := database.ConnectMSSQL()
    if err != nil {
        log.Fatal().Err(err).Msg("Failed to connect to MSSQL database")
    }
    defer func() {
        if err := msdb.Close(); err != nil {
            log.Error().Err(err).Msg("Error closing MSSQL database connection")
        }
    }()
    log.Info().Msg("Connected to MSSQL database successfully")
    router := routes.RegisterRoutes()
    router.HandleFunc("/actuator/info", HealthCheck).Methods("GET")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })
    handler := c.Handler(router)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }

    // Channel to listen for interrupt signals
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

    go func() {
        log.Info().Msg("Server started on port 8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal().Err(err).Msg("Failed to start server")
        }
    }()

    <-stop // Wait for the signal

    log.Info().Msg("Shutting down server...")
    // Create a context with a timeout for graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Attempt to gracefully shut down the server
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal().Err(err).Msg("Failed to shut down server")
    }
    log.Info().Msg("Server shut down successfully")
}
