package models

type TopSeller struct {
	ProductName string `json:"product_name"`
	TotalSold   int    `json:"total_sold"`
}
