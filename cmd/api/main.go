package main

import (
	"fmt"
	"os"
	config "home-market/internal/config"
)

func main() {
	fmt.Println("Hello, World!")

	//1. Load .env file
	config.LoadEnv()

	host := os.Getenv("DB_HOST")

    if host == "" {
        fmt.Println(".env gagal diload atau DB_HOST tidak ditemukan")
    } else {
        fmt.Println(".env berhasil diload. DB_HOST =", host)
    }

	//2. Connect to Database

	// Connect to PostgreSQL
	config.ConnectPostgres()
	defer config.PostgresDB.Close()
	fmt.Println("PostgreSQL connection closed.")

	// Connect to MongoDB
	config.ConnectMongo()
	fmt.Println("MongoDB connection closed.")
	

	
}