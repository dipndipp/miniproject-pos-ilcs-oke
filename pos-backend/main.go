package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"pos-backend/models"
	"strconv"

	"golang.org/x/crypto/bcrypt"

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
	func getProducts(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
	
		// Cek apakah data produk ada di Redis
		val, err := rdb.Get(ctx, "products").Result()
		if err == redis.Nil {
			// Data tidak ada di Redis, ambil dari database
			rows, err := db.Query("SELECT id, name, price, image_url FROM SYSBACKUP.PRODUCTS ORDER BY id ASC")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
	
			var products []models.Product
			for rows.Next() {
				var product models.Product
				err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.ImageURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
	
				base64Image, err := encodeImageToBase64(product.ImageURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				product.ImageURL = base64Image
				products = append(products, product)
			}
	
			if err = rows.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
	
			// Simpan data produk ke Redis dalam format JSON
			productsJSON, _ := json.Marshal(products)
			rdb.Set(ctx, "products", productsJSON, 0) // Set data di Redis tanpa expiration
	
			w.Header().Set("Content-Type", "application/json")
			w.Write(productsJSON) // Kirim data produk
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			// Jika data ada di Redis, kirimkan langsung
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(val)) // Kirim data produk dari cache
		}
	}

// encode image to base64
func encodeImageToBase64(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	imageData, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Encode ke base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	return "data:image/png;base64," + base64Image, nil
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
	

	//gambar
	func createProduct(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
	
		// Upload gambar
		imageURL, err := uploadImage(w, r)
		if err != nil {
			http.Error(w, "Image upload failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Decode JSON dari form untuk mengambil data produk
		var product models.Product
		product.Name = r.FormValue("name")
		product.Price, err = strconv.ParseFloat(r.FormValue("price"), 64)
		product.ImageURL = imageURL
	
		if err != nil {
			http.Error(w, "Invalid product data", http.StatusBadRequest)
			return
		}
	
		// Siapkan statement SQL untuk menghindari SQL injection
		stmt, err := db.Prepare("INSERT INTO SYSBACKUP.PRODUCTS (name, price, image_url) VALUES (:1, :2, :3) RETURNING id INTO :4")
		if err != nil {
			http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer stmt.Close()
	
		// Menyimpan ID produk terakhir yang dimasukkan
		var lastInsertID int
		_, err = stmt.Exec(product.Name, product.Price, product.ImageURL, &lastInsertID)
		if err != nil {
			http.Error(w, "Failed to execute statement: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Kirim respons sukses dengan ID produk yang baru ditambahkan
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Product added successfully with ID: %d", lastInsertID)
		rdb.Del(ctx, "products") // Menghapus cache Redis
	}
	
	// Update produk
	func updateProduct(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
	
		// Ambil ID dari URL
		id := r.URL.Path[len("/update-product/"):]
	
		// Debugging: Log ID yang diterima
		log.Println("Updating product with ID:", id)
	
		// Decode data produk dari form
		var product models.Product
		product.ID, _ = strconv.Atoi(id)
		product.Name = r.FormValue("name")
		product.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
	
		// Proses upload gambar (jika ada)
		var err error
		if r.MultipartForm != nil {
			imageURL, uploadErr := uploadImage(w, r)
			if uploadErr == nil {
				product.ImageURL = imageURL // Simpan URL gambar baru jika upload berhasil
			} else {
				http.Error(w, "Image upload failed: "+uploadErr.Error(), http.StatusInternalServerError)
				return
			}
		}
	
		// Debugging: Log product data
		log.Println("Product data to update:", product)
	
		// Query untuk update produk, gunakan gambar baru jika ada, jika tidak gunakan gambar lama
		query := `
			UPDATE SYSBACKUP.PRODUCTS 
			SET name = :1, price = :2, image_url = COALESCE(NULLIF(:3, ''), image_url) 
			WHERE id = :4
		`
		_, err = db.Exec(query, product.Name, product.Price, product.ImageURL, product.ID)
		if err != nil {
			http.Error(w, "Failed to update product: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Kosongkan cache di Redis setelah update
		rdb.Del(ctx, "products")
	
		// Kirim respons sukses
		fmt.Fprintf(w, "Product updated successfully with ID: %d", product.ID)
	}
	
	// Hapus produk
	func deleteProduct(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
	
		// Ambil ID dari URL
		id := r.URL.Path[len("/delete-product/"):]
	
		// Debugging: Log ID yang diterima
		log.Println("Deleting product with ID:", id)
	
		// Query untuk hapus produk
		_, err := db.Exec("DELETE FROM SYSBACKUP.PRODUCTS WHERE id = :1", id)
		if err != nil {
			http.Error(w, "Failed to delete product: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Kosongkan cache di Redis setelah penghapusan
		rdb.Del(ctx, "products")
	
		// Kirim respons sukses
		fmt.Fprintf(w, "Product deleted successfully with ID: %s", id)
	}
	func uploadImage(w http.ResponseWriter, r *http.Request) (string, error) {
		// Parse multipart form, allowing up to 10 MB files
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return "", fmt.Errorf("failed to parse form: %v", err)
		}
	
		// Retrieve the file from the form-data
		file, handler, err := r.FormFile("image")
		if err != nil {
			return "", fmt.Errorf("failed to retrieve file: %v", err)
		}
		defer file.Close()
	
		// Create a directory to store images, if not exists
		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	
		// Generate the file path to store the image
		filePath := filepath.Join("uploads", handler.Filename)
	
		// Create a new file in the uploads directory
		dst, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()
	
		// Copy the uploaded file data to the new file
		_, err = io.Copy(dst, file)
		if err != nil {
			return "", fmt.Errorf("failed to save file: %v", err)
		}
	
		// Return the file path as the image URL
		return filePath, nil
	}
	

	// Ambil semua pesanan dengan status 'On Progress'
	func getOrders(w http.ResponseWriter, r *http.Request) {
		// Query untuk mendapatkan pesanan yang statusnya 'On Progress'
		rows, err := db.Query("SELECT id, menu, status, total_price, created_at FROM SYSBACKUP.ORDERS WHERE status = 'On Progress' ORDER BY id ASC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
	
		var orders []models.Order
		for rows.Next() {
			var order models.Order
			var totalPrice sql.NullFloat64 // Gunakan sql.NullFloat64 untuk menangani nilai NULL
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
	
			// Query untuk mendapatkan detail pesanan dari tabel ORDER_DETAILS berdasarkan order_id
			detailRows, err := db.Query("SELECT product_name, quantity, total_price FROM SYSBACKUP.ORDER_DETAILS WHERE order_id = :1", order.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer detailRows.Close()
	
			// Menyimpan daftar detail pesanan ke dalam order
			var orderDetails []models.OrderDetail
			for detailRows.Next() {
				var detail models.OrderDetail
				err := detailRows.Scan(&detail.ProductName, &detail.Quantity, &detail.TotalPrice)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				orderDetails = append(orderDetails, detail)
			}
			order.Details = orderDetails
	
			orders = append(orders, order)
		}
	
		if err = rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Mengirimkan data dalam bentuk JSON
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
	
		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback() // Rollback if there's an error
	
		var orderID int
// Query INSERT dengan RETURNING, tanpa menggunakan sql.Named
query := `INSERT INTO SYSBACKUP.ORDERS (total_price, status) VALUES (:1, 'On Progress') RETURNING id INTO :2`
stmt, err := tx.Prepare(query)
if err != nil {
    http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
    return
}
defer stmt.Close()

// Menggunakan Exec untuk RETURNING INTO
_, err = stmt.Exec(order.TotalPrice, sql.Out{Dest: &orderID})
if err != nil {
    http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
    return
}
	
		// Iterate over items and insert into ORDER_DETAILS
		for _, item := range order.Items {
			// Cari product_id berdasarkan product_name
			var productID int
			err = tx.QueryRow("SELECT id FROM SYSBACKUP.PRODUCTS WHERE name = :1", item.ProductName).Scan(&productID)
			if err != nil {
				http.Error(w, "Failed to find product: "+err.Error(), http.StatusInternalServerError)
				return
			}
	
			// Insert ke ORDER_DETAILS dengan product_id
			detailQuery := "INSERT INTO SYSBACKUP.ORDER_DETAILS (order_id, product_name, quantity, total_price) VALUES (:1, :2, :3, :4)"
			_, err = tx.Exec(detailQuery, orderID, item.ProductName, item.Quantity, item.TotalPrice)
			if err != nil {
				http.Error(w, "Failed to create order details: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	
		// Commit transaction
		if err = tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
	
		// Respond with success
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Order created successfully"))
	}
	

	// Selesaikan pesanan
	// Complete Order
// Complete Order
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

	// Update status to 'Order Completed'
	queryComplete := "UPDATE SYSBACKUP.ORDERS SET status = 'Order Completed' WHERE id = :1"
	result, err := db.Exec(queryComplete, id)
	if err != nil {
		http.Error(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Optional: Check rows affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No order found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order marked as completed successfully. Order ID: " + id))
}

func cancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	// Update status to 'Order Canceled'
	queryCancel := "UPDATE SYSBACKUP.ORDERS SET status = 'Order Canceled' WHERE id = :1"
	result, err := db.Exec(queryCancel, id)
	if err != nil {
		http.Error(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Optional: Check rows affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No order found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order marked as canceled successfully. Order ID: " + id))
}


// Get Completed Orders
func getCompletedOrders(w http.ResponseWriter, r *http.Request) {
	// Query to get completed orders from the orders table
	rows, err := db.Query("SELECT id, menu, status, total_price, created_at FROM SYSBACKUP.ORDERS WHERE status IN ('Order Completed', 'Order Canceled') ORDER BY id ASC")
	if err != nil {
		http.Error(w, "Failed to retrieve completed orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var totalPrice sql.NullFloat64

		err := rows.Scan(&order.ID, &order.Menu, &order.Status, &totalPrice, &order.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to parse order data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if totalPrice.Valid {
			order.TotalPrice = &totalPrice.Float64
		}

		// Query to get the order details (products and quantity) for each order
		orderDetailsRows, err := db.Query("SELECT product_name, quantity FROM SYSBACKUP.ORDER_DETAILS WHERE order_id = :1", order.ID)
		if err != nil {
			http.Error(w, "Failed to retrieve order details: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer orderDetailsRows.Close()

		var details []models.OrderDetail
		for orderDetailsRows.Next() {
			var detail models.OrderDetail
			err := orderDetailsRows.Scan(&detail.ProductName, &detail.Quantity)
			if err != nil {
				http.Error(w, "Failed to parse order detail data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			details = append(details, detail)
		}

		order.Details = details // Add the details to the order
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error occurred during row iteration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		w.Write([]byte("No completed orders found"))
		return
	}

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


	func loginHandler(w http.ResponseWriter, r *http.Request) {
		var creds models.Credentials
	
		// Decode request body
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
	
		// Ambil user dari database
		user, err := getUserByUsername(creds.Username)
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
	func createAccount(w http.ResponseWriter, r *http.Request) {
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
func TopSeller(w http.ResponseWriter, r *http.Request) {
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
func TotalRevenue(w http.ResponseWriter, r *http.Request) {
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
func getProductList(w http.ResponseWriter, r *http.Request) {
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
func countOrderProgress(w http.ResponseWriter, r *http.Request) {
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
func countAdmin(w http.ResponseWriter, r *http.Request) {
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
func countCashier(w http.ResponseWriter, r *http.Request) {
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
		http.HandleFunc("/products", getProducts)
		http.HandleFunc("/product/", getProductByID) 
		http.HandleFunc("/create-product", createProduct)
		http.HandleFunc("/update-product/", updateProduct) 
		http.HandleFunc("/delete-product/", deleteProduct)


		// endpoint order
		http.HandleFunc("/orders", getOrders) // endpoint get all order
		http.HandleFunc("/create-order", createOrder) //endpoint buat order baru
		http.HandleFunc("/complete-order", completeOrder) // endpoint finish order
		http.HandleFunc("/cancel-order", cancelOrder) // endpoint finish order
		http.HandleFunc("/completed-orders", getCompletedOrders) // endpoint get all completed order
		http.HandleFunc("/delete-order", deleteOrder) //endpoint delete order



		// endpoint user
		http.HandleFunc("/login", loginHandler) // endpoint login
		http.HandleFunc("/create-account", createAccount)


		// endpoint dashboard
		http.HandleFunc("/top-selling-menu", TopSeller)
		http.HandleFunc("/total-revenue", TotalRevenue)
		http.HandleFunc("/product-count", getProductList)
		http.HandleFunc("/onprogress-count", countOrderProgress)
		http.HandleFunc("/admin-count", countAdmin)
		http.HandleFunc("/cashier-count", countCashier)

		
	
		// Use enableCORS for CORS handling
		log.Println("Server is running on port 8080...")
		if err := http.ListenAndServe(":8080", enableCORS(http.DefaultServeMux)); err != nil {
			log.Fatal("Error starting server:", err)
		}
	}
	

