package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var (
	db *sql.DB
	dbInitOnce sync.Once
)
func ConnectPostgres() (*sql.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Error().Msgf("error loading .env file: %v", err)
		return nil, err
	}
	log.Info().Msg("Loaded environment variables from .env file")
	log.Info().Msg("Connecting to PostgreSQL...")

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
				log.Info().Msgf("Attempt %d: Failed to connect to PostgreSQL: %v", i+1, err)
				time.Sleep(2 * time.Second)
				continue
			}

			time.Sleep(10 * time.Second)

			if err = db.Ping(); err == nil {
				break
			}
			log.Info().Msgf("Attempt %d: Failed to ping PostgreSQL: %v", i+1, err)
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			return
		}
		log.Info().Msg("Connected to PostgreSQL successfully")
		initPostgres()
	})

	return db, err
}

func initPostgres() {
	_, err := db.Exec("CALL create_pass_reset_logs_table()")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create password reset logs table")
	}
	log.Info().Msg("Password reset logs table created successfully")

	_, err = db.Exec("CALL create_admin_table()")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create admin table")
	}
	log.Info().Msg("Admin table created successfully")

	_, err = db.Exec("SELECT insert_into_admin($1, $2)", "admin", "admin123")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to insert values into the admin table")
	}
	log.Info().Msg("Values inserted into admin table successfully")
}
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal().Msg("Postgres Database connection is nil")
		log.Info().Msg("Initializing connection to PostgreSQL...")
		ConnectPostgres()
	}
	return db
}