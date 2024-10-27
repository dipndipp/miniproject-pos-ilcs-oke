package routes

import (
	"net/http"
	"pos-backend/handlers"
)

func RegisterUserRoutes() {
    http.HandleFunc("/login", handlers.LoginHandler)
    http.HandleFunc("/create-account", handlers.CreateAccount)
    http.HandleFunc("/count-admin", handlers.CountAdmin)
    http.HandleFunc("/count-cashier", handlers.CountCashier)
}