package models

type OrderDetail struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}
