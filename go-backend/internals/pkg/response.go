package pkg

import (
	"encoding/json"
	"log"
	"net/http"


)

func SendSuccessResponse(w http.ResponseWriter, message string)  {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
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

func SendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to send JSON response: %v", err)
	}
}

func SendErrorResponse2(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{"error": message}
	SendJSONResponse(w, response, statusCode)
}