package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-backend/internals/database"
	"go-backend/internals/pkg"
	"go-backend/models"
)

func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var request models.UpdatePasswordRequest
	var err error

	db := database.GetDB()
	msdb := database.GetMSDB()

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		pkg.SendErrorResponse(w, "Failed to decode request", http.StatusBadRequest)
		pkg.LogPasswordUpdate(db, "Password Update", request.Username, request.ServerIP, "Failed to decode request", err.Error())
		return
	}
	log.Printf("Received: Username: %s, ServerIP: %s, Email ID: %s", request.Username, request.ServerIP, request.Email)

	//create a log entry
	pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Pending", "Password update request received")
	log.Println("Password update request received")

	// validate the user credentials
	var isValidUser bool
	isValidUser, err = pkg.Check_user_credentials(msdb, request.Username, request.ServerIP, request.Email)
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to validate user credentials", http.StatusInternalServerError)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed to validate user credentials", err.Error())
		return
	}
	if !isValidUser {
		pkg.SendErrorResponse(w, "Invalid user credentials", http.StatusUnauthorized)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Invalid user credentials")
		return
	}
	log.Println("User credentials validated successfully")

	if request.OldPassword == request.NewPassword {
		log.Println("Old password and new password are the same")
		pkg.SendErrorResponse(w, "New password cannot be the same as the old password", http.StatusBadRequest)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "New password is the same as old password")
		return
	}

	isValid, err := pkg.CheckLoginExpiration(msdb, request.Username, request.ServerIP)
	if err != nil {
		log.Fatalf("Error checking login expiration: %v", err)
		pkg.SendErrorResponse(w, "Failed to check login expiration", http.StatusInternalServerError)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed to check login expiration", err.Error())
		return
	}

	if isValid {
		fmt.Println("Login is valid, proceeding to check old password via connecting...")
		// check if the old password is still valid
		isValidOldPassword, err := pkg.CheckOldPassword(msdb, request.Username, request.ServerIP, request.OldPassword, request.Database)
		if err != nil {
			pkg.SendErrorResponse(w, "Failed to check old password", http.StatusInternalServerError)
			pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed to check old password", err.Error())
			return
		}
		if !isValidOldPassword {
			fmt.Println("Old password is invalid.")
			pkg.SendErrorResponse(w, "Old password is invalid", http.StatusUnauthorized)
			pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Old password is invalid")
			return
		}
	} else {
		fmt.Println("Login is invalid or expired.")
		pkg.SendErrorResponse(w, "Login is invalid or expired", http.StatusUnauthorized)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Login is invalid or expired")
		return
	}
		
	// Calling the stored procedure to find related servers
	var serverReplicas []string 
	log.Println("Finding related server replicas...")
	serverReplicas, err = pkg.FindRelatedServers(msdb, request.ServerIP)
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to find related servers", http.StatusInternalServerError)
		pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Failed to find related servers", err.Error())
		return
	}
	log.Printf("Found related server replicas: %v", serverReplicas)
	// update password on the given server and related servers seperately by connecting to each server using admin credentials
	// update password on the given server
	_, err = msdb.Exec("EXEC dbo.ResetUserPassword @LoginName=?, @NewPassword=?, @OldPassword=?", request.Username, request.NewPassword, request.OldPassword)
    if err != nil {
        sqlErr := err.Error()
        pkg.SendErrorResponse(w, "Failed to update password on the server: "+sqlErr, http.StatusInternalServerError)
        pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed to update password on the server", sqlErr)
        return
    }
    log.Println("Password updated successfully on the server")


	// // update password on related servers
	for _, server := range serverReplicas {
		msWithUserCredstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
			server,
			request.Username,
			request.OldPassword,
			os.Getenv("MS_DB_PORT"),
			request.Database)
			
		msWithUserCred, err := sql.Open("mssql", msWithUserCredstr)
		if err != nil {
			pkg.SendErrorResponse(w, "Failed to connect to the related server", http.StatusInternalServerError)
			pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Failed to connect to the related server", err.Error())
			return
		}
		defer msWithUserCred.Close()

		if err = msWithUserCred.Ping(); err != nil {
			pkg.SendErrorResponse(w, "Failed to ping the related server", http.StatusInternalServerError)
			pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Failed to ping the related server", err.Error())
			return
		}
		log.Printf("Connected to the related server %s successfully", server)

	_, err = msWithUserCred.Exec("EXEC UpdatePassword @username=?, @newPassword=?, @oldpassword", request.Username, request.NewPassword, request.OldPassword)
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to update password on the related server", http.StatusInternalServerError)
		pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Failed to update password on the related server", err.Error())
		return
	}
	log.Printf("Password updated successfully on the related server %s", server)

	//update the access_requests table
	_, err = db.Exec("CALL update_pass_reset_logs($1, $2, $3, $4, $5)", request.Username, request.ServerIP, "Password Update", "Success", "Password updated successfully")
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to update pass_reset_logs table", http.StatusInternalServerError)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed to update access_requests table", err.Error())
		return
	}

	pkg.SendSuccessResponse(w, "Password updated successfully")
	// send email to the user once the password is updated
	err = pkg.SendConfirmationEmail(request.Email, request.Username)
	if err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
	}

	log.Printf("Password updated successfully for user: %s", request.Username)
}
}