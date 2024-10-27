package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"pos-backend/models"
	"strconv"

	_ "github.com/godror/godror"
)

var db *sql.DB
type User struct {
	Username string
	Password string
	    Role     string
}
func getProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, name, price FROM SYSBACKUP.PRODUCTS")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Price)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// Ambil produk berdasarkan ID
func getProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ambil ID dari URL path
	id := r.URL.Path[len("/product/"):]
	if id == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	var product models.Product
	query := "SELECT id, name, price FROM SYSBACKUP.PRODUCTS WHERE id = :1"
	err := db.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// Buat produk baru
func createProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON dari request body ke struct Product
	var product models.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi input produk
	if product.Name == "" || product.Price <= 0 {
		http.Error(w, "Invalid product data", http.StatusBadRequest)
		return
	}

	// Siapkan statement SQL untuk menghindari SQL injection
	stmt, err := db.Prepare("INSERT INTO SYSBACKUP.PRODUCTS (name, price) VALUES (:1, :2) RETURNING id INTO :3")
	if err != nil {
		http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Menyimpan ID produk terakhir yang dimasukkan
	var lastInsertID int
	_, err = stmt.Exec(product.Name, product.Price, &lastInsertID)
	if err != nil {
		http.Error(w, "Failed to execute statement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim respons sukses dengan ID produk yang baru ditambahkan
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Product added successfully with ID: %d", lastInsertID)
}

// Update produk
func updateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ambil ID dari URL
	id := r.URL.Path[len("/update-product/"):]

	var product models.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi input produk
	if product.Name == "" || product.Price <= 0 {
		http.Error(w, "Invalid product data", http.StatusBadRequest)
		return
	}

	product.ID, err = strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	query := "UPDATE SYSBACKUP.PRODUCTS SET name = :1, price = :2 WHERE id = :3"
	result, err := db.Exec(query, product.Name, product.Price, product.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product updated successfully"))
}

// Hapus produk berdasarkan ID
func deleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ambil ID dari URL path
	id := r.URL.Path[len("/delete-product/"):]
	if id == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM SYSBACKUP.PRODUCTS WHERE id = :1"
	_, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product deleted successfully"))
}

// Ambil semua pesanan dengan status 'On Progress'
func getOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, menu, status, total_price, created_at FROM SYSBACKUP.ORDERS WHERE status = 'On Progress' ORDER BY id ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var totalPrice sql.NullFloat64 // Use sql.NullFloat64 to handle NULL values
		err := rows.Scan(&order.ID, &order.Menu, &order.Status, &totalPrice, &order.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if totalPrice.Valid {
			order.TotalPrice = new(float64)
			*order.TotalPrice = totalPrice.Float64
		} else {
			order.TotalPrice = nil
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// Buat pesanan baru
func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "INSERT INTO SYSBACKUP.ORDERS (menu, total_price, status) VALUES (:1, :2, 'On Progress')"
	_, err = db.Exec(query, order.Menu, order.TotalPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Order created successfully"))
}

// Selesaikan pesanan
func completeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	// Update status menjadi 'Order Completed'
	queryComplete := "UPDATE SYSBACKUP.ORDERS SET status = 'Order Completed' WHERE id = :1"
	_, err := db.Exec(queryComplete, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order marked as completed successfully"))
}

// Get Completed Orders
func getCompletedOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, menu, status, total_price, created_at FROM SYSBACKUP.ORDERS WHERE status = 'Order Completed' ORDER BY id ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var totalPrice sql.NullFloat64
		err := rows.Scan(&order.ID, &order.Menu, &order.Status, &totalPrice, &order.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if totalPrice.Valid {
			order.TotalPrice = new(float64)
			*order.TotalPrice = totalPrice.Float64
		} else {
			order.TotalPrice = nil
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// Hapus order berdasarkan ID
func deleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM SYSBACKUP.ORDERS WHERE id = :1"
	_, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order deleted successfully"))
}

// buat auth/login
// Handler untuk login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Ambil user dari database
	user, err := getUserByUsername(creds.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Bandingkan password secara langsung
	if user.Password != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Login berhasil, kirim respons sukses dengan role
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"role":    user.Role, // Kirim role dalam response
	})
}
func getUserByUsername(username string) (*User, error) {
	var user User
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
