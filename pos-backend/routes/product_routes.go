package routes

import (
	"net/http"
	"pos-backend/handlers"
)

func RegisterProductRoutes() {
    http.HandleFunc("/products", handlers.GetProducts)
    http.HandleFunc("/product/", handlers.GetProductByID)
    http.HandleFunc("/create-product", handlers.CreateProduct)
    http.HandleFunc("/update-product/", handlers.UpdateProduct)
    http.HandleFunc("/delete-product/", handlers.DeleteProduct)
}