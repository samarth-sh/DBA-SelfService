package handlers

import (
	"encoding/json"
	"go-backend/internals/database"
	"go-backend/internals/pkg"
	"log"
	"net/http"


)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
    var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("Failed to decode admin login request: %v", err)
		pkg.SendErrorResponse(w, "Failed to decode admin login request", http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	pkg.HashPassword(credentials.Password)

	var isValidAdmin bool
	err := db.QueryRow("SELECT check_admin_credentials($1, $2)", credentials.Username, credentials.Password).Scan(&isValidAdmin)
	if err != nil {
		log.Printf("Failed to validate admin credentials: %v", err)
		pkg.SendErrorResponse(w, "Failed to validate admin credentials", http.StatusInternalServerError)
		return
	}

	if !isValidAdmin {
		log.Println("Invalid admin credentials provided")
		pkg.SendErrorResponse(w, "Invalid admin credentials", http.StatusUnauthorized)
		return
	}

	pkg.SendSuccessResponse(w, "Admin login successful")
	log.Println("Admin login successful")

	http.Redirect(w, r, "/getAllResetReq", http.StatusSeeOther)
}
