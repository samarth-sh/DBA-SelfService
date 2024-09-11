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
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Starting server...")

	initDB()
	defer db.Close()
	defer msdb.Close()
	defer msWithUserCred.Close()

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

var (
    db       *sql.DB
    msdb     *sql.DB
	msWithUserCred *sql.DB
    dbInitOnce sync.Once
	msdbInitOnce sync.Once
	msWithUserCredInitOnce sync.Once
)

func initDB() {
    if err := godotenv.Load(".env");
	err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
	log.Println("Loaded environment variables")
	log.Println("Connecting to PostgreSQL and MS SQL Server...")

	var err error
    dbInitOnce.Do(func() {
        pgconnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
            os.Getenv("DB_HOST"),
            os.Getenv("DB_PORT"),
            os.Getenv("DB_USER"),
            os.Getenv("DB_PASSWORD"),
            os.Getenv("DB_NAME"))

        for i := 0; i < 5; i++ {
            db, err = sql.Open("postgres", pgconnStr)
            if err != nil {
                log.Printf("Attempt %d: Failed to connect to PostgreSQL: %v", i+1, err)
                time.Sleep(2 * time.Second)
                continue
            }

            time.Sleep(10 * time.Second)

            if err = db.Ping(); err == nil {
                break
            }

            log.Printf("Attempt %d: Failed to ping PostgreSQL: %v", i+1, err)
            time.Sleep(2 * time.Second)
        }

        if err != nil {
            log.Fatalf("Could not connect to PostgreSQL: %v", err)
        }
        log.Println("Connected to PostgreSQL successfully")

        initPostgresDB()
    })

    msdbInitOnce.Do(func() {
        msconnStr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
            os.Getenv("MS_DB_SERVER"),
            os.Getenv("MS_DB_USER"),
            os.Getenv("MS_DB_PASSWORD"),
            os.Getenv("MS_DB_PORT"),
            os.Getenv("MS_DB_NAME"))

        msdb, err = sql.Open("mssql", msconnStr)
        if err != nil {
            log.Fatalf("Failed to connect to MS SQL Server: %v", err)
        }

        if err = msdb.Ping(); err != nil {
            log.Fatalf("Failed to ping MS SQL database: %v", err)
        }
        log.Println("Connected to MS SQL Server successfully")

        initMSSQLDB()
    })


}

func initPostgresDB() {
    _, err := db.Exec("CALL create_pass_reset_logs_table()")
    if err != nil {
        log.Fatal("Failed to create pass_reset_logs table: ", err)
    }
    log.Println("Password Reset Logs table created successfully")

    _, err = db.Exec("CALL create_admin_table()")
    if err != nil {
        log.Fatal("Failed to create admin table: ", err)
    }
    log.Println("Admin table created successfully")

    _, err = db.Exec("SELECT insert_into_admin($1, $2)", "admin", "admin123")
    if err != nil {
        log.Fatal("Failed to insert values into the admin table: ", err)
    }
    log.Println("Values inserted into the admin table successfully")
}

func initMSSQLDB() {
    initSQL, err := os.ReadFile("mssql_init.sql")
	log.Println("Reading MS SQL Server init script...")
    if err != nil {
        log.Fatal("Failed to read MS SQL Server init script: ", err)
    }

    _, err = msdb.Exec(string(initSQL))
	log.Println("Executing MS SQL Server init script...")
    if err != nil {
        log.Fatal("Failed to execute MS SQL Server init script: ", err)
    }
    log.Println("MS SQL Server initialization script executed successfully")
}

type UpdatePasswordRequest struct {
	Username    string `json:"username"`
	Email	   string `json:"emailID"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
	ServerIP    string `json:"serverIP"`
	Database	string `json:"database"`
}

func check_user_credentials(msdb *sql.DB, username, serverIP, emailID string) (bool, error) {
    var isValidUser bool
    log.Println("Checking user credentials...")
    log.Printf("Username: %s, ServerIP: %s, Email ID: %s", username, serverIP, emailID)

    query := `DECLARE @IsValid BIT;
              EXEC dbo.ValidateUserCredentials @Username = ?, @ServerIP = ?, @Email = ?, @IsValid = @IsValid OUTPUT;
              SELECT @IsValid;`

    log.Println("Executing query...")
    row := msdb.QueryRow(query, username, serverIP, emailID)
    log.Println("Query executed")
    log.Println("Scanning row...")
    if err := row.Scan(&isValidUser); err != nil {
        return false, err
    }
    log.Println("Row scanned")
    log.Println("User credentials checked successfully")

    return isValidUser, nil
}



// func findRelatedServers(msdb *sql.DB, serverIP string) ([]string, error) {
// 	query := "EXEC FindRelatedServers @ServerIP=?"
// 	rows, err := msdb.Query(query, serverIP)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var serverReplicas []string
// 	for rows.Next() {
// 		var serverIP string
// 		if err := rows.Scan(&serverIP); err != nil {
// 			return nil, err
// 		}
// 		serverReplicas = append(serverReplicas, serverIP)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return serverReplicas, nil
// }


func updatePassword(w http.ResponseWriter, r *http.Request) {
	var request UpdatePasswordRequest
	var err error

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
		logPasswordUpdate("Password Update", request.Username, request.ServerIP, "Failed to decode request", err.Error())
		return
	}
	log.Printf("Received: Username: %s, ServerIP: %s, Email ID: %s", request.Username, request.ServerIP, request.Email)

	//create a log entry
	logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Pending", "Password update request received")
	log.Println("Password update request received")

	// validate the user credentials
	var isValidUser bool
	isValidUser, err = check_user_credentials(msdb, request.Username, request.ServerIP, request.Email)
	if err != nil {
		sendErrorResponse(w, "Failed to validate user credentials", http.StatusInternalServerError)
		logStatus(request.Username, request.ServerIP, "Password Update", "Failed to validate user credentials", err.Error())
		return
	}
	if !isValidUser {
		sendErrorResponse(w, "Invalid user credentials", http.StatusUnauthorized)
		logStatus(request.Username, request.ServerIP, "Password Update", "Failed", "Invalid user credentials")
		return
	}
	log.Println("User credentials validated successfully")
	
	// using the user credentials to connect to the server
	msWithUserCredstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		request.ServerIP,
		request.Username,
		request.OldPassword,
		os.Getenv("MS_DB_PORT"),
		request.Database)
	
	msWithUserCredInitOnce.Do(func() {
		msWithUserCred, err = sql.Open("mssql", msWithUserCredstr)
		if err != nil {
			sendErrorResponse(w, "Failed to connect to the server using user provided credentials", http.StatusInternalServerError)
			logStatus(request.Username, request.ServerIP, "Password Update", "Failed to connect to the server using user provided", err.Error())
			return
		}
		defer msWithUserCred.Close()

		
		if err = msWithUserCred.Ping(); 
		err != nil {
			sendErrorResponse(w, "Failed to ping the server", http.StatusInternalServerError)
			logStatus(request.Username, request.ServerIP, "Password Update", "Failed to ping the server", err.Error())
			return
		}
		log.Println("Connected to the server successfully using user credentials")
	})

	// Calling the stored procedure to find related servers
	// var serverReplicas []string
	// log.Println("Finding related server replicas")
	// serverReplicas, err = findRelatedServers(msdb, request.ServerIP)
	// if err != nil {
	// 	sendErrorResponse(w, "Failed to find related servers", http.StatusInternalServerError)
	// 	logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Failed to find related servers", err.Error())
	// 	return
	// }
	// log.Printf("Found related server replicas: %v", serverReplicas)
	
	// update password on the given server and related servers seperately by connecting to each server using admin credentials
	// update password on the given server
	_, err = msdb.Exec("EXEC dbo.ResetUserPassword @LoginName=?, @NewPassword=?, @DisablePolicy=?, @DisableExpiration=?", request.Username, request.NewPassword, 1, 1)
	if err != nil {
		sendErrorResponse(w, "Failed to update password on the server", http.StatusInternalServerError)
		logStatus(request.Username, request.ServerIP, "Password Update", "Failed to update password on the server", err.Error())
		return
	}
	log.Println("Password updated successfully on the server")

	// update password on related servers
	// for _, server := range serverReplicas {
	// 	msWithAdminCredstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
	// 		server,
	// 		os.Getenv("MS_DB_USER"),
	// 		os.Getenv("MS_DB_PASSWORD"),
	// 		os.Getenv("MS_DB_PORT"),
	// 		request.Database)
	// 	msWithAdminCred, err := sql.Open("mssql", msWithAdminCredstr)
	// 	if err != nil {
	// 		sendErrorResponse(w, "Failed to connect to the related server", http.StatusInternalServerError)
	// 		logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Failed to connect to the related server", err.Error())
	// 		return
	// 	}
	// 	defer msWithAdminCred.Close()

	// 	if err = msWithAdminCred.Ping(); err != nil {
	// 		sendErrorResponse(w, "Failed to ping the related server", http.StatusInternalServerError)
	// 		logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Failed to ping the related server", err.Error())
	// 		return
	// 	}
	// 	log.Printf("Connected to the related server %s successfully", server)

		// _, err = msWithAdminCred.Exec("EXEC UpdatePassword @username=?, @newPassword=?, @oldpassword", request.Username, request.NewPassword, request.OldPassword)
		// if err != nil {
		// 	sendErrorResponse(w, "Failed to update password on the related server", http.StatusInternalServerError)
		// 	logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Failed to update password on the related server", err.Error())
		// 	return
		// }
		// log.Printf("Password updated successfully on the related server %s", server)

	//update the access_requests table
	_, err = db.Exec("CALL update_pass_reset_logs($1, $2, $3, $4, $5)", request.Username, request.ServerIP, "Password Update", "Success", "Password updated successfully")
	if err != nil {
		sendErrorResponse(w, "Failed to update pass_reset_logs table", http.StatusInternalServerError)
		logStatus(request.Username, request.ServerIP, "Password Update", "Failed to update access_requests table", err.Error())
		return
	}

	sendSuccessResponse(w, "Password updated successfully")
	
	logPasswordUpdate(request.Username, request.ServerIP, "Password Update", "Success", "Password updated successfully")
	log.Printf("Password updated successfully for user: %s", request.Username)
}
func logPasswordUpdate(username, serverIP, requestType, requestStatus, message string) {
	_, err := db.Exec("CALL log_updates($1, $2, $3, $4, $5)", username, serverIP, requestType, requestStatus, message)
	if err != nil {
		log.Printf("Failed to log password update: %v", err)
	}
}
func logStatus(username, serverIP, requestType, requestStatus, message string) {
	_, err := db.Exec("CALL update_pass_reset_logs($1, $2, $3, $4, $5)", username, serverIP, requestType, requestStatus, message)
	if err != nil {
		log.Printf("Failed to log status: %v", err)
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
	w.Header().Set("Content-Type", "application/json")
	if statusCode >= 400 {
		w.WriteHeader(statusCode)
	}
	response := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
	log.Printf("Error response sent: %v with status code %d", message, statusCode)
}
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func adminLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Failed to decode admin login request: %v", err)
		sendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
		return
	}

	hashPassword(credentials.Password)

	var isValidAdmin bool
	err := db.QueryRow("SELECT check_admin_credentials($1, $2)", credentials.Username, credentials.Password).Scan(&isValidAdmin)
	if err != nil {
		log.Printf("Failed to validate admin credentials: %v", err)
		sendErrorResponse(w, "Failed to validate admin credentials", http.StatusInternalServerError)
		return
	}

	if !isValidAdmin {
		log.Println("Invalid admin credentials provided")
		sendErrorResponse(w, "Invalid admin credentials", http.StatusUnauthorized)
		return
	}

	sendSuccessResponse(w, "Admin login successful")
	log.Println("Admin login successful")

	http.Redirect(w, r, "/getAllResetReq", http.StatusSeeOther)
}

type ResetRequest struct {
	RequestID      int    `json:"requestID"`
	Username       string `json:"username"`
	ServerIP       string `json:"serverIP"`
	RequestType    string `json:"requestType"`
	RequestStatus  string `json:"requestStatus"`
	RequestTime    string `json:"requestTime"`
}

func getAllResetReq(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM get_all_logs()")
	if err != nil {
		sendErrorResponse2(w, "Failed to query reset requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []ResetRequest
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		sendErrorResponse2(w, "Failed to load time zone", http.StatusInternalServerError)
		log.Printf("Failed to load time zone: %v", err)
		return
	}

	for rows.Next() {
		var request ResetRequest
		var requestTime pq.NullTime

		if requestTime.Valid {

			localTime := requestTime.Time.In(location)
			request.RequestTime = localTime.Format("2006-01-02 15:04:05")
		} else {
			request.RequestTime = requestTime.Time.Format("2006-01-02 15:04:05")
		}

		if err := rows.Scan(&request.RequestID, &request.Username, &request.ServerIP, &request.RequestType, &request.RequestStatus, &requestTime);
		err != nil {
			sendErrorResponse2(w, "Failed to scan reset requests", http.StatusInternalServerError)
			log.Printf("Failed to scan reset requests: %v", err)
			return
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		sendErrorResponse2(w, "Failed to iterate over reset requests", http.StatusInternalServerError)
		log.Printf("Failed to iterate over reset requests: %v", err)
		return
	}

	sendJSONResponse(w, requests, http.StatusOK)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Failed to send JSON response: %v", err)
    }
}

func sendErrorResponse2(w http.ResponseWriter, message string, statusCode int) {
    response := map[string]string{"error": message}
    sendJSONResponse(w, response, statusCode)
}