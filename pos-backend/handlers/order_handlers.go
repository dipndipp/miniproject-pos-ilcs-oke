package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"pos-backend/models"
	"strconv"

	"github.com/go-redis/redis/v8"
	_ "github.com/godror/godror"
)

// GetOrders retrieves orders with status 'On Progress'
func GetOrders(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, menu, status, total_price, created_at FROM SYSBACKUP.ORDERS WHERE status = 'On Progress' ORDER BY id ASC")
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

		detailRows, err := db.Query("SELECT product_name, quantity, total_price FROM SYSBACKUP.ORDER_DETAILS WHERE order_id = :1", order.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer detailRows.Close()

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// CreateOrder creates a new order
func CreateOrder(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var orderID int
	query := `INSERT INTO SYSBACKUP.ORDERS (total_price, status) VALUES (:1, 'On Progress') RETURNING id INTO :2`
	stmt, err := tx.Prepare(query)
	if err != nil {
		http.Error(w, "Failed to prepare statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(order.TotalPrice, sql.Out{Dest: &orderID})
	if err != nil {
		http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range order.Items {
		var productID int
		err = tx.QueryRow("SELECT id FROM SYSBACKUP.PRODUCTS WHERE name = :1", item.ProductName).Scan(&productID)
		if err != nil {
			http.Error(w, "Failed to find product: "+err.Error(), http.StatusInternalServerError)
			return
		}

		detailQuery := "INSERT INTO SYSBACKUP.ORDER_DETAILS (order_id, product_name, quantity, total_price) VALUES (:1, :2, :3, :4)"
		_, err = tx.Exec(detailQuery, orderID, item.ProductName, item.Quantity, item.TotalPrice)
		if err != nil {
			http.Error(w, "Failed to create order details: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Order created successfully. Order ID: " + strconv.Itoa(orderID)))
}

// CompleteOrder marks an order as completed
func CompleteOrder(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	queryComplete := "UPDATE SYSBACKUP.ORDERS SET status = 'Order Completed' WHERE id = :1"
	result, err := db.Exec(queryComplete, id)
	if err != nil {
		http.Error(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No order found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order marked as completed successfully. Order ID: " + id))
}

// CancelOrder marks an order as canceled
func CancelOrder(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	queryCancel := "UPDATE SYSBACKUP.ORDERS SET status = 'Order Canceled' WHERE id = :1"
	result, err := db.Exec(queryCancel, id)
	if err != nil {
		http.Error(w, "Failed to update order status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No order found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order marked as canceled successfully. Order ID: " + id))
}

// GetCompletedOrders retrieves completed orders
func GetCompletedOrders(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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

		order.Details = details
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

// DeleteOrder deletes an order by ID
func DeleteOrder(ctx context.Context, db *sql.DB, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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
	result, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, "Failed to delete order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No order found with the given ID", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order deleted successfully. Order ID: " + id))
}
