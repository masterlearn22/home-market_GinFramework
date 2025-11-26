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
	fmt.Println("Connected to PostgreSQL")
	fmt.Println("DB Postgresql :", os.Getenv("DB_NAME"))
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