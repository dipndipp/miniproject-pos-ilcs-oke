package routes

import (
	"context"
	"database/sql"
	"net/http"
	"pos-backend/handlers"

	"github.com/go-redis/redis/v8"
)
func RegisterProductRoutes(ctx context.Context, db *sql.DB, rdb *redis.Client) {



    http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetProducts(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/product/", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetProductByID(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/create-product", func(w http.ResponseWriter, r *http.Request) {
        handlers.CreateProduct(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/update-product/", func(w http.ResponseWriter, r *http.Request) {
        handlers.UpdateProduct(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/delete-product/", func(w http.ResponseWriter, r *http.Request) {
        handlers.DeleteProduct(ctx, db, rdb, w, r)
    })}