package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

func main() {
    log.Println("Starting server...")

    initDB()
    initAccessRequestDB()
    defer db.Close()
    defer accessRequestDB.Close()

    r := mux.NewRouter()

    r.HandleFunc("/update-password", updatePassword).Methods("PUT")
    r.HandleFunc("/access-request", accessRequest).Methods("POST")
    // r.HandleFunc("/update-password", func(w http.ResponseWriter, r *http.Request) {
    //     initDB()
    //     defer db.Close()
    //     updatePassword(w, r)
    // }).Methods("PUT")

    // // Route for access request
    // r.HandleFunc("/access-request", func(w http.ResponseWriter, r *http.Request) {
    //     initAccessRequestDB()
    //     defer accessRequestDB.Close()
    //     accessRequest(w, r)
    // }).Methods("POST")

    log.Println("Registered /update-password route and /access-request route")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })
    handler := c.Handler(r)

    log.Println("Server is running on port 8080")
    http.ListenAndServe(":8080", handler)
}

var db *sql.DB

func initDB() {
    var err error
    // Open the SQLite database located at "./data/users.db"
    db, err = sql.Open("sqlite3", "./data/passReset.db")
    if err != nil {
        log.Fatal("Failed to open database: ", err)
    }
    log.Println("Database opened successfully")

    // Create `users` table if it doesn't exist
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        serverIP TEXT NOT NULL
    )`)
    if err != nil {
        log.Fatal("Failed to create users table: ", err)
    }
    log.Println("Users table created successfully")

    // Insert initial values into `users` table
    _, err = db.Exec("INSERT INTO users (username, password, serverIP) VALUES (?, ?, ?)", "test_admin", "Test@dmin#45", "10.0.0.1")
    if err != nil {
        log.Fatal("Failed to insert values into the users table: ", err)
    }
    log.Println("Values inserted into the users table successfully")

    // Create `logs` table if it doesn't exist
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (
        request_id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        serverIP TEXT NOT NULL,
        request_type TEXT DEFAULT 'Password Update',
        request_status TEXT DEFAULT 'Pending',
        message TEXT,
        request_time DATETIME DEFAULT (DATETIME('now', 'localtime'))
    )`)
    if err != nil {
        log.Fatal("Failed to create logs table: ", err)
    }
    log.Println("Logs table created successfully")
}
var accessRequestDB *sql.DB

func initAccessRequestDB() {
    var err error
    accessRequestDB, err = sql.Open("sqlite3", "./data/accessRequest.db")
    if err != nil {
        log.Fatalf("Failed to connect to access request database: %v", err)
    }

    // Create `access_requests` table
    _, err = accessRequestDB.Exec(`
    DROP TABLE IF EXISTS access_requests;
        CREATE TABLE IF NOT EXISTS access_requests (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            database TEXT NOT NULL,
            request_type TEXT NOT NULL,
            request_status TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Fatalf("Failed to create access_requests table: %v", err)
    }
    log.Println("access_requests table created successfully")

    // Create `access_logs` table
    _, err = accessRequestDB.Exec(`
    DROP TABLE IF EXISTS access_logs;
        CREATE TABLE IF NOT EXISTS access_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        request_id INTEGER NOT NULL,
        action TEXT NOT NULL,
        username TEXT NOT NULL,
        database TEXT NOT NULL,
        access_level TEXT NOT NULL,
        status TEXT NOT NULL,
        message TEXT NOT NULL,
        log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(request_id) REFERENCES access_requests(id)
        )
    `)
    if err != nil {
        log.Fatalf("Failed to create access_logs table: %v", err)
    }
    log.Println("access_logs table created successfully")

    log.Println("Access control database initialized with tables: access_requests, access_logs")
}
type UpdatePasswordRequest struct {
    Username string `json:"username"`
    OldPassword string `json:"oldPassword"`
    NewPassword string `json:"newPassword"`
    ServerIP string `json:"serverIP"`
}
type AccessRequest struct {
    Username    string `json:"username"`
    Database    string `json:"database"`
    AccessLevel string `json:"accessLevel"`
    Reason      string `json:"reason"`
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
    var request UpdatePasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        sendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to decode request", err.Error())
        return
    }
    log.Printf("Received: Username: %s, ServerIP: %s", request.Username, request.ServerIP)

    //Check if the user exists in the database and retrieve the password
    var checkPass string
    log.Println("Checking if user exists in the database")
    err := db.QueryRow("SELECT password FROM users WHERE username = ?", request.Username).Scan(&checkPass)
    if err == sql.ErrNoRows {
        sendErrorResponse(w, "User not found", http.StatusNotFound)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: User not found", "User does not exist in the database")
        return
    } else if err != nil {
        sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
        return
    }
    log.Println("User exists in the database")

    log.Printf("Password in the database: %s", checkPass)
    log.Printf("Checking if password matches the one in the database")
    //Check if password matches the one in the database
    if request.OldPassword != checkPass {
        sendErrorResponse(w, "Invalid password", http.StatusUnauthorized)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid password", "Current password does not match the one in the database")
        return
    }
    log.Println("Password matches the one in the database")

    log.Println("Checking if the server IP matches the one in the database")
    //Check if the server IP matches the one in the database
    var checkServerIP string
    err = db.QueryRow("SELECT serverIP FROM users WHERE username = ?", request.Username).Scan(&checkServerIP)
    if err != nil {
        sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
        return
    }
    if checkServerIP != request.ServerIP {
        sendErrorResponse(w, "Invalid server IP", http.StatusBadRequest)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid server IP", "Server IP does not match the one in the database")
        return
    }
    log.Println("Server IP matches the one in the database")

    log.Println("Updating password")
    //Update password
    newPass := request.NewPassword
    _, err = db.Exec("UPDATE users SET password = ? WHERE username = ? AND serverIP = ?", newPass, request.Username, request.ServerIP)
    if err != nil {
        sendErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to update password", err.Error())
        return
    }
    log.Println("Password updated")
    sendSuccessResponse(w, "Password updated")
    logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Success", "Password updated successfully")
}
func logPasswordUpdate(requestType, username, serverIP, requestStatus, message string) {
    _, err := db.Exec("INSERT INTO logs (request_type, username, serverIP, request_status, message) VALUES (?, ?, ?, ?, ?)", requestType, username, serverIP, requestStatus, message)
    if err != nil {
        log.Printf("Failed to log password update: %v", err)
    }
}


func accessRequest(w http.ResponseWriter, r *http.Request) {
    var request AccessRequest
    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil {
        sendErrorResponse(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    log.Printf("Received Access Request: Username: %s, Database: %s, Access Level: %s, Reason: %s", 
    request.Username, request.Database, request.AccessLevel, request.Reason)
    
    // Process the access request (save to database)
    requestID, err := saveAccessRequest(request)
    if err != nil {
        sendErrorResponse(w, "Failed to process access request", http.StatusInternalServerError)
        logAccessRequest(requestID, "Access Request", request.Username, request.Database, request.AccessLevel, "Failure", "Access request could not be submitted")
        return
    }
    log.Printf("Access request submitted successfully")
    logAccessRequest(requestID, "Access Request", request.Username, request.Database, request.AccessLevel, "Success", "Access request submitted successfully")
    
    sendSuccessResponse(w, "Access request submitted")
}

func saveAccessRequest(request AccessRequest) (int64, error) {
    result, err := accessRequestDB.Exec(`
        INSERT INTO access_requests (username, database, request_type, request_status) 
        VALUES (?, ?, ?, ?)`,
        request.Username, request.Database, request.AccessLevel, "Pending")
    if err != nil {
        log.Printf("Error inserting into access_requests: %v", err)
        return 0, err
    }
    
    requestID, err := result.LastInsertId()
    if err != nil {
        log.Printf("Error getting last insert ID: %v", err)
        return 0, err
    }

    return requestID, nil
}

func logAccessRequest(requestID int64, action, username, database, accessLevel, status, message string) {
    _, err := accessRequestDB.Exec(`
        INSERT INTO access_logs (request_id, action, username, database, access_level, status, message, log_time) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        requestID, action, username, database, accessLevel, status, message, time.Now())
    if err != nil {
        log.Println("Failed to log access request:", err)
    }
}

func sendSuccessResponse(w http.ResponseWriter, message string) {
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": message})
}
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
        log.Printf("Failed to send error response: %v", err)
    }
    log.Printf("Error response sent: %v with status code %d", message, statusCode)
}