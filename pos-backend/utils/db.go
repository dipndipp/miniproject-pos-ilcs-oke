package utils

import (
	"database/sql"
	"log"

	_ "github.com/godror/godror"
)

var DB *sql.DB

func InitDB() {
    var err error
    DB, err = sql.Open("godror", "system/123456789@//localhost:1521/orc1")
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }
}