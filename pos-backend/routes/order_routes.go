package routes

import (
	"net/http"
	"pos-backend/handlers"
)

func RegisterOrderRoutes() {
    http.HandleFunc("/orders", handlers.GetOrders)
    http.HandleFunc("/create-order", handlers.CreateOrder)
    http.HandleFunc("/complete-order", handlers.CompleteOrder)
    http.HandleFunc("/cancel-order", handlers.CancelOrder)
    http.HandleFunc("/completed-orders", handlers.GetCompletedOrders)
    http.HandleFunc("/delete-order", handlers.DeleteOrder)
}