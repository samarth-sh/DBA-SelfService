package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/rs/zerolog/log"
)

var (
	msdb *sql.DB
	msdbInitOnce sync.Once
)

func ConnectMSSQL() (*sql.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	log.Info().Msg("Loaded environment variables from .env file")
	log.Info().Msg("Connecting to MS SQL Server...")

	var err error
	msdbInitOnce.Do(func() {
		msconnStr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
			os.Getenv("MS_DB_SERVER"),
			os.Getenv("MS_DB_USER"),
			os.Getenv("MS_DB_PASSWORD"),
			os.Getenv("MS_DB_PORT"),
			os.Getenv("MS_DB_NAME"))

		msdb, err = sql.Open("mssql", msconnStr)
		if err != nil {
			return
		}

		if err = msdb.Ping(); err != nil {
			return
		}
		log.Info().Msg("Connected to MS SQL Server successfully")

		initMSSQL()
	})

	return msdb, err
}
func initMSSQL() {
	directories := []string{"MSSQL-SP", "MSSQL-UDF"}

	for _, dir := range directories {
		log.Info().Msgf("Executing initialization scripts from directory: %s", dir)

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(d.Name()) == ".sql" {
				// log.Printf("Reading file: %s", path)

				sqlContent, err := os.ReadFile(path)
				if err != nil {
					log.Fatal().Err(err).Msgf("Failed to read SQL script from file: %s", path)
				}

				log.Printf("Executing SQL script from: %s", path)
				_, err = msdb.Exec(string(sqlContent))
				if err != nil {
					log.Fatal().Err(err).Msgf("Failed to execute SQL script from: %s", path)
				}
				// log.Printf("Successfully executed SQL script from: %s", path)
			}

			return nil
		})

		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to read initialization scripts from directory: %s", dir)
		}
	}
	log.Info().Msg("Successfully executed all initialization scripts")
}
func GetMSDB () *sql.DB {
	if msdb == nil {
		log.Fatal().Msg("MS SQL Server connection is not initialized")
		log.Info().Msg("Initializing MS SQL Server connection...")
		ConnectMSSQL()
	}

	return msdb
}