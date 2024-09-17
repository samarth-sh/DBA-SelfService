package routes

import (
    "github.com/gorilla/mux"
    "go-backend/internals/handlers"
)

func RegisterRoutes() *mux.Router {
    r := mux.NewRouter()

    r.HandleFunc("/update-password", handlers.UpdatePassword).Methods("PUT")
    r.HandleFunc("/admin-login", handlers.AdminLogin).Methods("POST")
    r.HandleFunc("/getAllResetReq", handlers.GetAllResetReq).Methods("GET")

    return r
}
