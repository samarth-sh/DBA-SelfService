package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	// "regexp"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
    // "golang.org/x/crypto/bcrypt"
)

type UpdatePasswordRequest struct {
    Username string `json:"username"`
    OldPassword string `json:"oldPassword"`
    NewPassword string `json:"newPassword"`
    ServerIP string `json:"serverIP"`
}
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
        log.Printf("Failed to send error response: %v", err)
    }
    log.Printf("Error response sent: %v with status code %d", message, statusCode)
}

func sendSuccessResponse(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
    var request UpdatePasswordRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        //http.Error(w, err.Error(), http.StatusBadRequest)
        sendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to decode request", err.Error())
        return
    }

    // // Check if the server IP is valid and of the format 10.xxx.xxx.xxx
    // if !regexp.MustCompile(`^10\.\d{1,3}\.\d{1,3}\.\d{1,3}$`).MatchString(request.ServerIP) {
    //    // http.Error(w, "Invalid server IP", http.StatusBadRequest)
    //     sendErrorResponse(w, "Invalid server IP", http.StatusBadRequest)
    //     logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid server IP", "Server IP is not in the format/does not match any known server")
    //     return
    // }

    //Check if the user exists in the database and retrieve the password
    var checkPass string
    err := db.QueryRow("SELECT password FROM users WHERE username = ?", request.Username).Scan(&checkPass)
    if err == sql.ErrNoRows {
        // http.Error(w, "User not found", http.StatusNotFound)
        sendErrorResponse(w, "User not found", http.StatusNotFound)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: User not found", "User does not exist in the database")
        return
    } else if err != nil {
        // http.Error(w, err.Error(), http.StatusInternalServerError)
        sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
        return
    }

    //Check if password matches the one in the database
    if request.OldPassword != checkPass {
        // http.Error(w, "Invalid password", http.StatusUnauthorized)
        sendErrorResponse(w, "Invalid password", http.StatusUnauthorized)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid password", "Old password does not match the one in the database")
        return
    }

    //Check if the server IP matches the one in the database
    var checkServerIP string
    err = db.QueryRow("SELECT serverIP FROM users WHERE username = ?", request.Username).Scan(&checkServerIP)
    if err != nil {
        //http.Error(w, err.Error(), http.StatusInternalServerError)
        sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
        return
    }
    if checkServerIP != request.ServerIP {
        // http.Error(w, "Invalid server IP", http.StatusBadRequest)
        sendErrorResponse(w, "Invalid server IP", http.StatusBadRequest)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid server IP", "Server IP does not match the one in the database")
        return
    }

    //Update password
    newPass := request.NewPassword
    _, err = db.Exec("UPDATE users SET password = ? WHERE username = ? AND serverIP = ?", newPass, request.Username, request.ServerIP)
    if err != nil {
        // http.Error(w, err.Error(), http.StatusInternalServerError)
        sendErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
        logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to update password", err.Error())
        return
    }
    sendSuccessResponse(w, "Password updated")
    logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Success", "Password updated successfully")
}


var db *sql.DB
var logDB *sql.DB

func main() {

    r := mux.NewRouter()
    r.HandleFunc("/update-password", updatePassword).Methods("PUT")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })
    handler := c.Handler(r)

    initDB()
    initLogDB()
    defer db.Close()

    log.Println("Server is running on port 8080")
    http.ListenAndServe(":8080", handler)
}

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./users.db")
    if err != nil {
        log.Fatal(err)
    }
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        password TEXT NOT NULL,
        serverIP TEXT NOT NULL
    )`)
    if err != nil {
        log.Fatal(err)
    }
}
func initLogDB() {
    var err error
    logDB, err = sql.Open("sqlite3", "./logs.db")
    if err != nil {
        log.Fatal(err)
    }
    _, err = logDB.Exec(`CREATE TABLE IF NOT EXISTS logs (
        request_id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        serverIP TEXT NOT NULL,
        request_type Text DEFAULT 'Password Update',
        request_status Text DEFAULT 'Pending',
        message TEXT,
        request_time DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        log.Fatal(err)
    }
}
func logPasswordUpdate(requestType, username, serverIP, requestStatus, message string) {
    _, err := logDB.Exec("INSERT INTO logs (request_type, username, serverIP, request_status, message) VALUES (?, ?, ?, ?, ?)", requestType, username, serverIP, requestStatus, message)
    if err != nil {
        log.Printf("Failed to log password update: %v", err)
    }
}