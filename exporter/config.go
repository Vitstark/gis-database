package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Config holds the application configuration
type Config struct {
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	OutputFile       string
}

// ConnectPostgreSQL connects to the PostgreSQL database
func ConnectPostgreSQL(cfg Config) (*sql.DB, error) {
	pgDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)

	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// CreateGeoPackage creates a new GeoPackage file and returns a database connection
func CreateGeoPackage(outputFile string) (*sql.DB, error) {
	// Remove existing GeoPackage file if it exists
	if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing file: %w", err)
	}

	db, err := sql.Open("sqlite3", outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create GeoPackage: %w", err)
	}

	return db, nil
}

// CloseDB safely closes a database connection
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}
