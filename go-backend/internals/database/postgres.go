package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
	dbInitOnce sync.Once
)
func ConnectPostgres() (*sql.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	log.Println("Loaded environment variables")
	log.Println("Connecting to PostgreSQL...")

	var err error
	dbInitOnce.Do(func() {
		pgconnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"))

		for i := 0; i < 5; i++ {
			db, err = sql.Open("postgres", pgconnStr)
			if err != nil {
				log.Printf("Attempt %d: Failed to connect to PostgreSQL: %v", i+1, err)
				time.Sleep(2 * time.Second)
				continue
			}

			time.Sleep(10 * time.Second)

			if err = db.Ping(); err == nil {
				break
			}

			log.Printf("Attempt %d: Failed to ping PostgreSQL: %v", i+1, err)
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			return
		}
		log.Println("Connected to PostgreSQL successfully")
		initPostgres()
	})

	return db, err
}

func initPostgres() {
	_, err := db.Exec("CALL create_pass_reset_logs_table()")
	if err != nil {
		log.Fatal("Failed to create pass_reset_logs table: ", err)
	}
	log.Println("Password Reset Logs table created successfully")

	_, err = db.Exec("CALL create_admin_table()")
	if err != nil {
		log.Fatal("Failed to create admin table: ", err)
	}
	log.Println("Admin table created successfully")

	_, err = db.Exec("SELECT insert_into_admin($1, $2)", "admin", "admin123")
	if err != nil {
		log.Fatal("Failed to insert values into the admin table: ", err)
	}
	log.Println("Values inserted into the admin table successfully")
}
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("Database connection is not initialized")
		log.Println("Connecting to PostgreSQL...")
		ConnectPostgres()
	}
	return db
}