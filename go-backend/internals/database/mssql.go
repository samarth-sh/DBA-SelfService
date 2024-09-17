package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"


	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
)

var (
	msdb *sql.DB
	msdbInitOnce sync.Once
)

func ConnectMSSQL() (*sql.DB, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	log.Println("Loaded environment variables")
	log.Println("Connecting to MS SQL Server...")

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
		log.Println("Connected to MS SQL Server successfully")

		initMSSQL()
	})

	return msdb, err
}
func initMSSQL() {
	directories := []string{"MSSQL-SP", "MSSQL-UDF"}

	for _, dir := range directories {
		log.Printf("Reading SQL files from directory: %s", dir)

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(d.Name()) == ".sql" {
				log.Printf("Reading file: %s", path)

				sqlContent, err := os.ReadFile(path)
				if err != nil {
					log.Fatalf("Failed to read file: %s, error: %v", path, err)
				}

				log.Printf("Executing SQL script from: %s", path)
				_, err = msdb.Exec(string(sqlContent))
				if err != nil {
					log.Fatalf("Failed to execute SQL script from file: %s, error: %v", path, err)
				}
				log.Printf("Successfully executed SQL script from: %s", path)
			}

			return nil
		})

		if err != nil {
			log.Fatalf("Failed to read files from directory %s: %v", dir, err)
		}
	}

	log.Println("MS SQL Server initialization scripts executed successfully")

}
func GetMSDB () *sql.DB {
	if msdb == nil {
		log.Fatal("MSSQL is not connected")
		log.Println("Connecting to MS SQL Server...")
		ConnectMSSQL()
	}

	return msdb
}