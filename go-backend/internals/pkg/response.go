package pkg

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
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
		log.Error().Msgf("Failed to send error response: %v", err)
	}
	log.Info().Msgf("Error response sent: %v with status code %d", message, statusCode)
}

func SendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to send JSON response")
	}
}

func SendErrorResponse2(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{"error": message}
	SendJSONResponse(w, response, statusCode)
}