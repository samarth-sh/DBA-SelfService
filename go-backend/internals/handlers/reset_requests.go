package handlers

import (
	"go-backend/internals/database"
	"go-backend/internals/pkg"
	"go-backend/models"
	"log"
	"net/http"

	"github.com/lib/pq"
)


func GetAllResetReq(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT * FROM get_all_logs()")
	if err != nil {
			pkg.SendErrorResponse2(w, "Failed to query reset requests", http.StatusInternalServerError)
			return
	}
	defer rows.Close()
	
	var requests []models.ResetRequest
	for rows.Next() {
		var request models.ResetRequest
		var requestTime pq.NullTime
	
		if err := rows.Scan(&request.RequestID, &request.Username, &request.ServerIP, &request.RequestType, &request.RequestStatus, &request.Message, &requestTime); err != nil {
			pkg.SendErrorResponse(w, "Failed to scan reset requests", http.StatusInternalServerError)
			log.Printf("Failed to scan reset requests: %v", err)
			return
		}
	
		if requestTime.Valid {
			request.RequestTime = requestTime.Time.Format("2006-01-02 15:04:05")
		} else {
			request.RequestTime = "N/A"
		}
	
		requests = append(requests, request)
	}
	
	if err := rows.Err(); err != nil {
		pkg.SendErrorResponse2(w, "Failed to iterate over reset requests", http.StatusInternalServerError)
		log.Printf("Failed to iterate over reset requests: %v", err)
		return
	}
	
	pkg.SendJSONResponse(w, requests, http.StatusOK)
	}
	

