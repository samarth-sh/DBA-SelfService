package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

var (
    db               *sql.DB
    dbInitOnce       sync.Once
)

func main() {
    log.Println("Starting server...")

    initDB()
    defer db.Close()

    r := mux.NewRouter()

    r.HandleFunc("/update-password", updatePassword).Methods("PUT")
    r.HandleFunc("/admin-login", adminLogin).Methods("POST")
    r.HandleFunc("/getAllResetReq", getAllResetReq).Methods("GET")

    log.Println("Registered routes")

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

func initDB() {
    dbInitOnce.Do(func() {
        var err error
        db, err = sql.Open("sqlite3", "./data/passReset.db")
        if err != nil {
            log.Fatal("Failed to open database: ", err)
        }
        log.Println("Database opened successfully")

        _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            password TEXT NOT NULL,
            serverIP TEXT NOT NULL,
            read_permission INTEGER DEFAULT 0,
            write_permission INTEGER DEFAULT 0,
            admin_permission INTEGER DEFAULT 0
        )`)
        if err != nil {
            log.Fatal("Failed to create users table: ", err)
        }
        log.Println("Users table created successfully")

        _, err = db.Exec("INSERT INTO users (username, password, serverIP) VALUES (?, ?, ?)", "test_admin", "Test@dmin#45", "10.0.0.1")
        if err != nil {
            log.Fatal("Failed to insert values into the users table: ", err)
        }
        log.Println("Values inserted into the users table successfully")

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
        
    })
}


type UpdatePasswordRequest struct {
    Username string `json:"username"`
    OldPassword string `json:"oldPassword"`
    NewPassword string `json:"newPassword"`
    ServerIP string `json:"serverIP"`
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
func adminLogin(w http.ResponseWriter, r *http.Request) {
    var credentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        sendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
        log.Printf("Failed to decode admin login request: %v", err)
        return
    }

    // Check if the credentials are valid
    if credentials.Username != "admin" || credentials.Password != "admin123" {
        sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
        log.Printf("Invalid admin login credentials: %s", credentials.Username)
        return
    }

    // Successful login, call getAllResetReq
    getAllResetReq(w, r)
}

func getAllResetReq(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT * FROM logs")
    if err != nil {
        sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
        log.Printf("Failed to query logs table: %v", err)
        return
    }
    defer rows.Close()

    var requests []map[string]interface{}
    for rows.Next() {
        var (
            requestID    int
            username     string
            serverIP     string
            requestType  string
            requestStatus string
            message      string
            requestTime  string
        )
        if err := rows.Scan(&requestID, &username, &serverIP, &requestType, &requestStatus, &message, &requestTime); err != nil {
            sendErrorResponse(w, "Failed to scan row", http.StatusInternalServerError)
            log.Printf("Failed to scan row: %v", err)
            return
        }
        request := map[string]interface{}{
            "requestID":    requestID,
            "username":     username,
            "serverIP":     serverIP,
            "requestType":  requestType,
            "requestStatus": requestStatus,
            "message":      message,
            "requestTime":  requestTime,
        }
        requests = append(requests, request)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(requests)
}
