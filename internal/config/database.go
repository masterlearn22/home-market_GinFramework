package config

import (
	"log"
	"os"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq" 
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
	"time"
)

var PostgresDB *sql.DB
func ConnectPostgres() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	PostgresDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	var host, port, dbname, user, schema string

	PostgresDB.QueryRow("SELECT inet_server_addr(), inet_server_port(), current_database(), current_user, current_schema()").
		Scan(&host, &port, &dbname, &user, &schema)

	fmt.Println("🔥 Go connected to:")
	fmt.Println("Host  :", host)
	fmt.Println("Port  :", port)
	fmt.Println("DB    :", dbname)
	fmt.Println("User  :", user)
	fmt.Println("Schema:", schema)

}

var MongoDB *mongo.Database
func ConnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	MongoDB = client.Database(os.Getenv("MONGO_DB_NAME"))
	fmt.Println("Connected to MongoDB")
	fmt.Println("DB MongoDB :", os.Getenv("MONGO_DB_NAME"))
	
}