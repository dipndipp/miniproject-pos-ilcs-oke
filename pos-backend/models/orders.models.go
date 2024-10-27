package models

import "time"

type Order struct {
    ID        int          `json:"id"`
    Menu      string       `json:"menu"`
    Status    string       `json:"status"`
    CreatedAt time.Time    `json:"created_at"`
    TotalPrice *float64    `json:"total_price"`
	Items      []OrderItem   `json:"items"` // List of ordered items
    Details    []OrderDetail `json:"details"` 
}

type OrderItem struct {
    ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
    TotalPrice  float64 `json:"total_price"`
}