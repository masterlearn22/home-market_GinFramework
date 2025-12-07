package main

import (
	"fmt"
	config "home-market/internal/config"
	_ "database/sql"
	"home-market/internal/delivery/http/route"
)

// @title           Home Market API
// @version         1.0
// @description     This is the API documentation for the Home Market project.
// @host      localhost:8080
// @BasePath  /api
// @insert "database/sql"
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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
	mongoClient := config.MongoDB.Client()
	
	//3. Setup Gin App
	var app = config.SetupGin()

	//4. Initialize Routes
	route.SetupRoute(app, config.PostgresDB, mongoClient)
	fmt.Println("Setup route berhasil")

	//5. Run the server
	config.SetupServer(app)
}