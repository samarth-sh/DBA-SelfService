package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go-backend/internals/database"
	"go-backend/internals/pkg"
	"go-backend/models"
	"github.com/rs/zerolog/log"
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
	log.Info().Msgf("Received password update request for user: %s, Email: %s, serverIP: %s", request.Username, request.Email, request.ServerIP)

	//create a log entry
	pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Pending", "Password update request received")
	log.Info().Msg("Password update request received")

	// validate the user credentials
	var isValidUser bool
	isValidUser, err = pkg.Check_user_credentials(msdb, request.Username, request.ServerIP, request.Email)
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to validate user credentials", http.StatusInternalServerError)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Pending: Failed to validate user credentials", err.Error())
		return
	}
	if !isValidUser {
		pkg.SendErrorResponse(w, "Invalid user credentials", http.StatusUnauthorized)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Invalid user credentials")
		return
	}
	log.Info().Msg("User credentials validated successfully")

	if request.OldPassword == request.NewPassword {
		log.Info().Msg("New password cannot be the same as the old password")
		pkg.SendErrorResponse(w, "New password cannot be the same as the old password", http.StatusBadRequest)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Pending: New password is the same as old password")
		return
	}

	isValid, err := pkg.CheckLoginExpiration(msdb, request.Username, request.ServerIP)
	if err != nil {
		log.Info().Msgf("Error checking login expiration: %v", err)
		pkg.SendErrorResponse(w, "Failed to check login existence", http.StatusInternalServerError)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Pending: Failed to check login expiration", err.Error())
		return
	}

	if isValid {
		log.Info().Msg("Login is still valid, checking old password...")
		// check if the old password is still valid
		isValidOldPassword, err := pkg.CheckOldPassword(msdb, request.Username, request.ServerIP, request.OldPassword, request.Database)
		if err != nil {
			pkg.SendErrorResponse(w, "Failed to check old password", http.StatusInternalServerError)
			pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Pending: Failed to check old password", err.Error())
			return
		}
		if !isValidOldPassword {
			log.Info().Msg("Old password is invalid")
			pkg.SendErrorResponse(w, "Old password is invalid", http.StatusUnauthorized)
			pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Pending: Old password is invalid")
			return
		}
	} else {
		log.Info().Msg("Login is invalid or expired")
		pkg.SendErrorResponse(w, "Login is invalid or expired", http.StatusUnauthorized)
		pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Failed", "Pending: Login is invalid or expired")
		return
	}
		
	// Calling the stored procedure to find related servers
	var serverReplicas []string 
	log.Info().Msg("Finding related server replicas...")
	serverReplicas, err = pkg.FindRelatedServers(msdb, request.ServerIP)
	if err != nil {
		pkg.SendErrorResponse(w, "Failed to find related servers", http.StatusInternalServerError)
		pkg.LogPasswordUpdate(db, request.Username, request.ServerIP, "Password Update", "Pending: Failed to find related servers", err.Error())
		return
	}
	log.Info().Msgf("Related server replicas found: %v", serverReplicas)
	// update password on the given server and related servers seperately by connecting to each server using admin credentials
	// update password on the given server
	_, err = msdb.Exec("EXEC dbo.ResetUserPassword @LoginName=?, @NewPassword=?, @DisablePolicy=?, @DisableExpiration=?", request.Username, request.NewPassword, 1,1)
    if err != nil {
        sqlErr := err.Error()
        pkg.SendErrorResponse(w, "Failed to update password on the server ", http.StatusInternalServerError)
        pkg.LogStatus(db, request.Username, request.ServerIP, "Password Update", "Pending: Failed to update password on the server", sqlErr)
        return
    }
    log.Info().Msg("Password updated successfully on the given server")


	// // update password on related servers
	for _, server := range serverReplicas {
		msdb.SetConnMaxLifetime(1)
		newConnstr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
			server,
			os.Getenv("MS_DB_USER"),
			os.Getenv("MS_DB_PASSWORD"),
			os.Getenv("MS_DB_PORT"),
			os.Getenv("MS_DB_NAME"))
		msdb, err = sql.Open("mssql", newConnstr)
		if err != nil {
			pkg.SendErrorResponse(w, "Failed to connect to the server", http.StatusInternalServerError)
			pkg.LogStatus(db, request.Username, server, "Password Update", "Pending: Failed to connect to the server", err.Error())
			return
		}
		defer msdb.Close()
		_, err = msdb.Exec("EXEC dbo.ResetUserPassword @LoginName=?, @NewPassword=?, @DisablePolicy=?, @DisableExpiration=?", request.Username, request.NewPassword, 1,1)
		if err != nil {
			sqlErr := err.Error()
			pkg.SendErrorResponse(w, "Failed to update password on the server ", http.StatusInternalServerError)
			pkg.LogStatus(db, request.Username, server, "Password Update", "Pending: Failed to update password on the server", sqlErr)
			return
		}
		log.Info().Msgf("Password updated successfully on the related server: %s", server)
	}

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
		log.Error().Err(err).Msg("Failed to send confirmation email")
	}

	log.Info().Msg("Password updated successfully for the user: " + request.Username)
}
