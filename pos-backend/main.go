package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"pos-backend/routes"

	"github.com/go-redis/redis/v8"
	_ "github.com/godror/godror"
)


var (
	db     *sql.DB
	ctx    = context.Background()
	rdb    *redis.Client
)
	type User struct {
		Username string
		Password string
			Role     string
	}

	func enableCORS(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			// Handle preflight request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
	
			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
	




	// function utama
	func main() {
		var err error
		db, err = sql.Open("godror", "system/123456789@//localhost:1521/orc1") // Sesuaikan koneksi database
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer db.Close()
		rdb = redis.NewClient(&redis.Options{
			Addr: "localhost:6379", // Ganti dengan alamat Redis Anda
		})
		// printHashedPassword("kasirpassword") // Ganti dengan password yang ingin diuji

		// endpoint produk
		

        routes.RegisterProductRoutes(ctx, db, rdb)
		// endpoint order
		routes.RegisterOrderRoutes(ctx, db, rdb)


		// endpoint user
		routes.RegisterUserRoutes(ctx, db, rdb)

		
	
		// Use enableCORS for CORS handling
		log.Println("Server is running on port 8080...")
		if err := http.ListenAndServe(":8080", enableCORS(http.DefaultServeMux)); err != nil {
			log.Fatal("Error starting server:", err)
		}
	}
	