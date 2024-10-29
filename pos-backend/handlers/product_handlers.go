package handlers

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

	"github.com/go-redis/redis/v8"
	_ "github.com/godror/godror"
)


func GetProducts(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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
func GetProductByID(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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
func CreateProduct(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Upload gambar
	imageURL, err := uploadImage(ctx, db, rdb, w, r)
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
func UpdateProduct(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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
		imageURL, uploadErr := uploadImage(ctx, db, rdb, w, r)
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
func DeleteProduct(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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
func uploadImage(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) (string, error) {
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
