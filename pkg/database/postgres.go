package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func Connect() *sql.DB {
	err := godotenv.Load()

	if err != nil {
		log.Println("Warning: .env not found")
	}

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatalf("Warning: DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to the database!")

	return db
}