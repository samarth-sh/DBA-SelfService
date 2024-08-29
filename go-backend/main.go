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

		var err error

		for i := 0; i < 5; i++ {
			db, err = sql.Open("postgres", connStr)
			if err != nil {
				log.Printf("Attempt %d: Failed to connect to database: %v", i+1, err)
				time.Sleep(2 * time.Second)
				continue
			}

			time.Sleep(10 * time.Second)

			err = db.Ping()
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

		_, err = db.Exec("CALL create_users_table()")
		if err != nil {
			log.Fatal("Failed to create users table: ", err)
		}
		log.Println("Users table created successfully")

		_, err = db.Exec("CALL insert_into_users($1, $2, $3)", "testt_user", "Test@userfortest123", "10.1.1.1")
		if err != nil {
			log.Fatal("Failed to insert values into the users table: ", err)
		}
		log.Println("Values inserted into the users table successfully")

		_, err = db.Exec("CALL create_logs_table()")
		if err != nil {
			log.Fatal("Failed to create logs table: ", err)
		}
		log.Println("Logs table created successfully")

		_, err = db.Exec("CALL create_admin_table()")
		if err != nil {
			log.Fatal("Failed to create admin table: ", err)
		}
		log.Println("Admin table created successfully")

		_, err = db.Exec("CALL insert_into_admin($1, $2)", "admin", "admin123")
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

	// Check if the user exists in the database and retrieve the password
	
	var userExists bool
	log.Println("Checking if user exists in the database")
	//run the stored procedure to check if the user exists in the database which returns boolean value
	err := db.QueryRow("SELECT user_exists($1)", request.Username).Scan(&userExists)
	if err != nil {
		sendErrorResponse(w, "Failed to query database - usercheck", http.StatusInternalServerError)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
		return
	}
	if !userExists {
		sendErrorResponse(w, "User not found", http.StatusNotFound)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: User not found", "User does not exist in the database")
		return
	}
	var checkPass string
	err = db.QueryRow("SELECT get_user_password($1)", request.Username).Scan(&checkPass)
	if err == sql.ErrNoRows {
		sendErrorResponse(w, "User not found/password could not be retrieved(verified)", http.StatusNotFound)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: User not found", "Password could not be retrieved(verified)")
		return
	} else if err != nil {
		sendErrorResponse(w, "Failed to query database - usercheck", http.StatusInternalServerError)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
		return
	}
	log.Println("User exists in the database")

	log.Println("Checking if the server IP matches the one in the database")
	// Check if the server IP matches the one in the database
	var checkServerIP string
	err = db.QueryRow("SELECT get_serverip($1)", request.Username).Scan(&checkServerIP)
	if err != nil {
		sendErrorResponse(w, "Failed to query database - serverip", http.StatusInternalServerError)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to query database", err.Error())
		return
	}
	if checkServerIP != request.ServerIP {
		sendErrorResponse(w, "Invalid server IP", http.StatusBadRequest)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid server IP", "Server IP does not match the one in the database")
		return
	}
	log.Println("Server IP matches the one in the database")


	log.Printf("Password in the database: %s", checkPass)
	log.Printf("Provided password: %s", request.OldPassword)
	log.Println("Checking if password matches the one in the database")

	// Check if password matches the one in the database
// 	if checkPass != request.OldPassword {
// 		sendErrorResponse(w, "Invalid password", http.StatusUnauthorized)
// 		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid password", "Current password does not match the one in the database")
// 		return
// 	}
// 	log.Println("Password matches the one in the database")

// 	//Insert the new password into the database
// 	_, err = db.Exec("CALL update_user_password($1, $2, $3)", request.NewPassword, request.Username, request.ServerIP)
// 	if err != nil {
// 		sendErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
// 		log.Printf("Failed to update password: %v", err)
// 		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to update password", err.Error())
// 		return
// 	}

// 	sendSuccessResponse(w, "Password updated successfully")
// 	logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Success", "Password updated successfully")
// 	log.Printf("Password updated successfully for user: %s", request.Username)
// }

	if err := bcrypt.CompareHashAndPassword([]byte(checkPass), []byte(request.OldPassword)); err != nil {
		log.Printf("Password comparison failed: %v", err)
		sendErrorResponse(w, "Invalid password", http.StatusUnauthorized)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed: Invalid password", "Current password does not match the one in the database")
		return
	}

	log.Println("Password matches the one in the database")

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		sendErrorResponse(w, "Failed to hash password", http.StatusInternalServerError)
		log.Printf("Failed to hash new password: %v", err)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to hash password", err.Error())
		return
	}

	// Update password with the hashed value
	_, err = db.Exec("CALL update_user_password($1, $2, $3)", hashedPassword, request.Username, request.ServerIP)
	if err != nil {
		sendErrorResponse(w, "Failed to update password", http.StatusInternalServerError)
		log.Printf("Failed to update password: %v", err)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to update password", err.Error())
		return
	}

	sendSuccessResponse(w, "Password updated successfully")
	logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Success", "Password updated successfully")
	log.Printf("Password updated successfully for user: %s", request.Username)
}

func logPasswordUpdate(requestType, username, serverIP, requestStatus, message string) {
	_, err := db.Exec("CALL log_updates($1, $2, $3, $4 ,$5)", requestType, username, serverIP, requestStatus, message)
	if err != nil {
		log.Printf("Failed to log password update: %v", err)
	}
}

func sendSuccessResponse(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
// 		log.Printf("Failed to send error response: %v", err)
// 	}
// 	log.Printf("Error response sent: %v with status code %d", message, statusCode)
// }
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	// Set Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Write the status code only if it hasn't been written yet
	if statusCode >= 400 {
		w.WriteHeader(statusCode)
	}

	// Encode the error message as JSON
	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log if encoding fails but don't write another response
		log.Printf("Failed to send error response: %v", err)
	}
	
	// Log the error response for debugging purposes
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
	var isValidAdmin bool
	err := db.QueryRow("SELECT check_admin_credentials($1, $2)", credentials.Username, credentials.Password).Scan(&isValidAdmin)
	if err != nil {
		sendErrorResponse(w, "Failed to validate admin credentials", http.StatusInternalServerError)
		log.Printf("Failed to validate admin credentials: %v", err)
		return
	}

	if !isValidAdmin {
		sendErrorResponse(w, "Invalid admin credentials", http.StatusUnauthorized)
		log.Println("Invalid admin credentials provided")
		return
	}

	sendSuccessResponse(w, "Admin login successful")
	log.Println("Admin login successful")
	getAllResetReq(w, r)
}

func getAllResetReq(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM get_all_logs()")
	if err != nil {
		sendErrorResponse(w, "Failed to query reset requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		sendErrorResponse(w, "Failed to load time zone", http.StatusInternalServerError)
		log.Printf("Failed to load time zone: %v", err)
		return
	}

	for rows.Next() {
		var requestID int
		var requestType, username, serverIP, requestStatus, message string
		var requestTime pq.NullTime

		if err := rows.Scan(&requestID, &requestType, &username, &serverIP, &requestStatus, &message, &requestTime); err != nil {
			sendErrorResponse(w, "Failed to scan reset requests", http.StatusInternalServerError)
			log.Printf("Failed to scan reset requests: %v", err)
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
		"requestID":      requestID,
		"username":       username,
		"serverIP":       serverIP,
		"requestType":    requestType,
		"requestStatus":  requestStatus,
		"message":        message,
		"requestTime":    formattedRequestTime,
	}
	requests = append(requests, request)
}

if err := rows.Err(); err != nil {
	sendErrorResponse(w, "Failed to iterate over reset requests", http.StatusInternalServerError)
	log.Printf("Failed to iterate over reset requests: %v", err)
	return
}

if len(requests) == 0 {
	sendSuccessResponse(w, "No reset requests found")
	return
}else {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

}

