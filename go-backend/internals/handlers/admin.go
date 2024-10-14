package handlers

import (
	"encoding/json"
	"net/http"

	"go-backend/internals/database"
	"go-backend/internals/pkg"
	"github.com/rs/zerolog/log"

)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
    var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Info().Msgf("Failed to decode admin login request: %v", err)
		pkg.SendErrorResponse(w, "Failed to decode admin login request", http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	pkg.HashPassword(credentials.Password)

	var isValidAdmin bool
	err := db.QueryRow("SELECT check_admin_credentials($1, $2)", credentials.Username, credentials.Password).Scan(&isValidAdmin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate admin credentials")
		pkg.SendErrorResponse(w, "Failed to validate admin credentials", http.StatusInternalServerError)
		return
	}

	if !isValidAdmin {
		log.Info().Msg("Invalid admin credentials")
		pkg.SendErrorResponse(w, "Invalid admin credentials", http.StatusUnauthorized)
		return
	}

	// pkg.SendSuccessResponse(w, "Admin login successful")
	log.Info().Msg("Admin login successful")

	http.Redirect(w, r, "/getAllResetReq", http.StatusSeeOther)
}
