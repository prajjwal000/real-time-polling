package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func Connect() *sql.DB {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "polluser")
	password := getEnv("POSTGRES_PASSWORD", "pollpass")
	dbname := getEnv("POSTGRES_DB", "polling")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var db *sql.DB
	var err error

	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("Failed to open database: %v", err)
			time.Sleep(time.Second)
			continue
		}
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Database not ready (attempt %d/30): %v", i+1, err)
		time.Sleep(time.Second)
	}

	if err != nil {
		log.Fatalf("Could not connect to database after 30 attempts: %v", err)
	}

	log.Println("Connected to PostgreSQL")
	return db
}

func RunMigrations(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS polls (
		id SERIAL PRIMARY KEY,
		question TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS options (
		id SERIAL PRIMARY KEY,
		poll_id INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
		text TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS votes (
		id SERIAL PRIMARY KEY,
		poll_id INTEGER NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
		option_id INTEGER NOT NULL REFERENCES options(id) ON DELETE CASCADE,
		ip_address VARCHAR(45) NOT NULL,
		timestamp TIMESTAMP DEFAULT NOW(),
		UNIQUE(poll_id, ip_address)
	);
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations complete")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
