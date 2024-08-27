package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var (
	db         *sql.DB
	dbInitOnce sync.Once
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
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	// var db *sql.DB
	var err error

	for i := 0; i < 5; i++ { 
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Attempt %d: Failed to connect to database: %v", i+1, err)
			time.Sleep(2 * time.Second) // wait before retrying
			continue
		}

		err = db.Ping() // verify the connection
		if err == nil {
			break
		}

		log.Printf("Attempt %d: Failed to ping database: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	log.Println("Connected to the database successfully")

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
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

		_, err = db.Exec("INSERT INTO users (username, password, serverIP) VALUES ($1, $2, $3)", "test_admin", "Test@dmin#45", "10.0.0.1")
		if err != nil {
			log.Fatal("Failed to insert values into the users table: ", err)
		}
		log.Println("Values inserted into the users table successfully")

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (
			request_id SERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			serverIP TEXT NOT NULL,
			request_type TEXT DEFAULT 'Password Update',
			request_status TEXT DEFAULT 'Pending',
			message TEXT,
			request_time TIMESTAMPTZ DEFAULT NOW()
		)`)
		if err != nil {
			log.Fatal("Failed to create logs table: ", err)
		}
		log.Println("Logs table created successfully")

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS admin (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			password TEXT NOT NULL
		)`)
		if err != nil {
			log.Fatal("Failed to create admin table: ", err)
		}
		log.Println("Admin table created successfully")

		_, err = db.Exec("INSERT INTO admin (username, password) VALUES ($1, $2)", "admin", "admin123")
		if err != nil {
			log.Fatal("Failed to insert values into the admin table: ", err)
		}
		log.Println("Values inserted into the admin table successfully")
	})
}

type UpdatePasswordRequest struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	ServerIP    string `json:"serverIP"`
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
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", request.Username).Scan(&checkPass)
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
	err = db.QueryRow("SELECT serverIP FROM users WHERE username = $1", request.Username).Scan(&checkServerIP)
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

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		sendErrorResponse(w, "Failed to hash password", http.StatusInternalServerError)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to hash password", err.Error())
		return
	}

	// Update password with the hashed value
	_, err = db.Exec("UPDATE users SET password = $1 WHERE username = $2 AND serverIP = $3", hashedPassword, request.Username, request.ServerIP)
	if err != nil {
		sendErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to update password", err.Error())
		return
	}
	sendSuccessResponse(w, "Password updated successfully")
	logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Success", "Password updated successfully")

	log.Printf("Password updated successfully for user: %s", request.Username)
}

func logPasswordUpdate(requestType, username, serverIP, requestStatus, message string) {
	_, err := db.Exec("INSERT INTO logs (request_type, username, serverIP, request_status, message) VALUES ($1, $2, $3, $4, $5)", requestType, username, serverIP, requestStatus, message)
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

	// Check if the admin credentials are valid
	var adminusername string
	if err := db.QueryRow("SELECT username FROM admin").Scan(&adminusername); err != nil {
		sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query admin username: %v", err)
		return
	}
	
	var adminPassword string
	if err := db.QueryRow("SELECT password FROM admin").Scan(&adminPassword); err != nil {
		sendErrorResponse(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query admin password: %v", err)
		return
	}
	
	if credentials.Username != adminusername || credentials.Password != adminPassword {
		sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		err := fmt.Errorf("invalid credentials")
		log.Printf("Invalid credentials: %v", err)
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
    // time zone offset for india
    location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		sendErrorResponse(w, "Failed to load time zone", http.StatusInternalServerError)
		log.Printf("Failed to load time zone: %v", err)
		return
	}

	for rows.Next() {
		var requestID int
		var username, serverIP, requestType, requestStatus, message string
		var requestTime pq.NullTime

		if err := rows.Scan(&requestID, &username, &serverIP, &requestType, &requestStatus, &message, &requestTime); err != nil {
			sendErrorResponse(w, "Failed to scan row", http.StatusInternalServerError)
			log.Printf("Failed to scan row: %v", err)
			return
		}

        var formattedRequestTime string
		if requestTime.Valid {
			localTime := requestTime.Time.In(location)
			formattedRequestTime = localTime.Format("2006-01-02 15:04:05") // Custom format
		} else {
			formattedRequestTime = "N/A"
		}

		request := map[string]interface{}{
			"requestID":     requestID,
			"username":      username,
			"serverIP":      serverIP,
			"requestType":   requestType,
			"requestStatus": requestStatus,
			"message":       message,
			"requestTime":   formattedRequestTime,
		}
		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		sendErrorResponse(w, "Error occurred during iteration", http.StatusInternalServerError)
		log.Printf("Error during rows iteration: %v", err)
		return
	}

	if len(requests) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Nothing to display"})
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(requests)
	}
}


