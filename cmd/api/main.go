package main

import (
	"fmt"
	config "home-market/internal/config"
	"home-market/internal/delivery/http/route"
)

func main() {
	// fmt.Println("Hello, World!")

	//1. Load .env file
	config.LoadEnv()

	//2. Connect to Database

	// Connect to PostgreSQL
	config.ConnectPostgres()
	defer config.PostgresDB.Close()

	// Connect to MongoDB
	config.ConnectMongo()
	
	//3. Setup Gin App
	var app = config.SetupGin()

	//4. Initialize Routes
	route.SetupPostgres(app, config.PostgresDB)
	fmt.Println("Setup route berhasil")

	//5. Run the server
	config.SetupServer(app)


}