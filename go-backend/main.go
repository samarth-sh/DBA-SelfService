package main

import (
    "log"
    "net/http"
    "go-backend/internals/database"
    "go-backend/routes"
    "github.com/rs/cors"
)

func main() {
    log.Println("Starting server...")

    db, err := database.ConnectPostgres()
    if err != nil {
        log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
    }
    defer func() {
        if err := db.Close(); err != nil {
            log.Printf("Error closing PostgreSQL database connection: %v", err)
        }
    }()
    log.Println("Connected to PostgreSQL database successfully")

    msdb, err := database.ConnectMSSQL()
    if err != nil {
        log.Fatalf("Failed to connect to MSSQL database: %v", err)
    }
    defer func() {
        if err := msdb.Close(); err != nil {
            log.Printf("Error closing MSSQL database connection: %v", err)
        }
    }()
    log.Println("Connected to MSSQL database successfully")

    router := routes.RegisterRoutes()    

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })
    handler := c.Handler(router)

    log.Println("Server is running on port 8080")
    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
