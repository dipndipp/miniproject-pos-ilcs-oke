package routes

import (
	"context"
	"database/sql"
	"net/http"
	"pos-backend/handlers"

	"github.com/go-redis/redis/v8"
)

    func RegisterOrderRoutes(ctx context.Context, db *sql.DB, rdb *redis.Client) {
        
 
        http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
            handlers.GetOrders(ctx, db, rdb, w, r)
        })
        http.HandleFunc("/create-order", func(w http.ResponseWriter, r *http.Request) {
            handlers.CreateOrder(ctx, db, rdb, w, r)
        })
        http.HandleFunc("/complete-order", func(w http.ResponseWriter, r *http.Request) {
            handlers.CompleteOrder(ctx, db, rdb, w, r)
        })
        http.HandleFunc("/cancel-order", func(w http.ResponseWriter, r *http.Request) {
            handlers.CancelOrder(ctx, db, rdb, w, r)
        })
        http.HandleFunc("/completed-orders", func(w http.ResponseWriter, r *http.Request) {
            handlers.GetCompletedOrders(ctx, db, rdb, w, r)
        })
        http.HandleFunc("/delete-order", func(w http.ResponseWriter, r *http.Request) {
            handlers.DeleteOrder(ctx, db, rdb, w, r)
        })
    }
    
  