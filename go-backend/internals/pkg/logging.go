package pkg

import (
	"log"
	"database/sql"


)

func LogPasswordUpdate(db *sql.DB, username, serverIP, requestType, requestStatus, message string) {
	_, err := db.Exec("CALL log_updates($1, $2, $3, $4, $5)", username, serverIP, requestType, requestStatus, message)
	if err != nil {
		log.Printf("Failed to log password update: %v", err)
	}	
}
func LogStatus(db *sql.DB, username, serverIP, requestType, requestStatus, message string) {
	_, err := db.Exec("CALL update_pass_reset_logs($1, $2, $3, $4, $5)", username, serverIP, requestType, requestStatus, message)
	if err != nil {
		log.Printf("Failed to log status: %v", err)
	}
}


