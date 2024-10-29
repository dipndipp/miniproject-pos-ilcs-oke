package routes

import (
	"context"
	"database/sql"
	"net/http"
	"pos-backend/handlers"

	"github.com/go-redis/redis/v8"
)

func RegisterUserRoutes(ctx context.Context, db *sql.DB, rdb *redis.Client) {

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        handlers.LoginHandler(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/create-account", func(w http.ResponseWriter, r *http.Request) {
        handlers.CreateAccount(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/count-admin", func(w http.ResponseWriter, r *http.Request) {
        handlers.CountAdmin(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/count-cashier", func(w http.ResponseWriter, r *http.Request) {
        handlers.CountCashier(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/top-selling-menu", func(w http.ResponseWriter, r *http.Request) {
        handlers.TopSeller(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/total-revenue", func(w http.ResponseWriter, r *http.Request) {
        handlers.TotalRevenue(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/product-count", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetProductList(ctx, db, rdb, w, r)
    })
    http.HandleFunc("/onprogress-count", func(w http.ResponseWriter, r *http.Request) {
        handlers.CountOrderProgress(ctx, db, rdb, w, r)
    })


}


 