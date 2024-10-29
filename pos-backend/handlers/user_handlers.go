package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"pos-backend/models"

	"github.com/go-redis/redis/v8"
	_ "github.com/godror/godror"
	"golang.org/x/crypto/bcrypt"
)


func LoginHandler(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Ambil user dari database
	user, err := getUserByUsername(creds.Username, db)
	if err != nil || user == nil {
		http.Error(w, "Username atau Password salah", http.StatusUnauthorized)
		return
	}

	// Cek log untuk hash yang diambil dari database
	fmt.Println("Hash dari DB:", user.Password)

	// Bandingkan password hash di database dengan password yang dimasukkan
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Password Salah", http.StatusUnauthorized)
		return
	}

	// Jika login berhasil, kirim respons sukses dengan role
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"role":    user.Role,
	})
}

func HashPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
func CreateAccount(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Hash password yang akan disimpan
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Log untuk memastikan hash telah berhasil dibuat
	fmt.Println("Hash baru yang akan disimpan:", string(hashedPassword))

	// Insert user ke database
	query := "INSERT INTO SYSBACKUP.USERS (username, password, role) VALUES (:1, :2, :3)"
	_, err = db.Exec(query, user.Username, hashedPassword, user.Role)
	if err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account created successfully"})
}
func getUserByUsername(username string,db *sql.DB ) (*models.User, error) {
	var user models.User
	query := "SELECT username, password, role FROM SYSBACKUP.USERS WHERE username = :1" // Sesuaikan dengan nama tabel Anda
	err := db.QueryRow(query, username).Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User tidak ditemukan
		}
		return nil, err // Error lain
	}
	return &user, nil // Kembalikan user
}

// func hashPassword(password string) (string, error) {
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(hashedPassword), nil
// }
// func printHashedPassword(password string) {
// 	hashedPassword, err := hashPassword(password)
// 	if err != nil {
// 		log.Fatalf("Error hashing password: %v", err)
// 	}
// 	fmt.Println("Hashed password:", hashedPassword)
// 	fmt.Println("Password:", password)

// }

// buat dashboard
// Function untuk showcase menu paling laris
func TopSeller(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
query := `
	SELECT product_name, SUM(quantity) as total_sold
	FROM SYSBACKUP.ORDER_DETAILS
	GROUP BY product_name
	ORDER BY total_sold DESC
	FETCH FIRST 1 ROWS ONLY
`
rows, err := db.Query(query)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}
defer rows.Close()

var topSellingProducts []models.TopSeller
for rows.Next() {
	var topProduct models.TopSeller
	err := rows.Scan(&topProduct.ProductName, &topProduct.TotalSold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topSellingProducts = append(topSellingProducts, topProduct)
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(topSellingProducts)
}

// Function untuk mendapatkan total pendapatan
func TotalRevenue(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
var totalRevenue float64
query := `
	SELECT COALESCE(SUM(total_price), 0) 
	FROM SYSBACKUP.ORDERS 
	WHERE status = 'Order Completed'
`

err := db.QueryRow(query).Scan(&totalRevenue)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]float64{"total_revenue": totalRevenue})
}


// Function untuk mendapatkan daftar produk
func GetProductList(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
var count int
query := "SELECT COUNT(*) FROM SYSBACKUP.PRODUCTS"

err := db.QueryRow(query).Scan(&count)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"product_count": count})
}
func CountOrderProgress(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
var count int
query := "SELECT COUNT(*) FROM SYSBACKUP.ORDERS WHERE status = 'On Progress'"

err := db.QueryRow(query).Scan(&count)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"order_onprogress_count": count})
}
func CountAdmin(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
var count int
query := "SELECT COUNT(*) FROM SYSBACKUP.USERS WHERE role = 'admin'"

err := db.QueryRow(query).Scan(&count)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"admin_count": count})
}
func CountCashier(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
var count int
query := "SELECT COUNT(*) FROM SYSBACKUP.USERS WHERE role = 'kasir'"

err := db.QueryRow(query).Scan(&count)
if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"cashier_count": count})
}
