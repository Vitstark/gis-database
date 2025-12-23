package main

import (
	"flag"
	"log"
)

func main() {
	var (
		pgHost     = flag.String("pg-host", "localhost", "PostgreSQL host")
		pgPort     = flag.Int("pg-port", 5432, "PostgreSQL port")
		pgUser     = flag.String("pg-user", "postgres", "PostgreSQL user")
		pgPassword = flag.String("pg-password", "postgres", "PostgreSQL password")
		pgDB       = flag.String("pg-db", "postgres", "PostgreSQL database name")
		outputFile = flag.String("output", "cadastral.gpkg", "Output GeoPackage file path")
	)
	flag.Parse()

	cfg := Config{
		PostgresHost:     *pgHost,
		PostgresPort:     *pgPort,
		PostgresUser:     *pgUser,
		PostgresPassword: *pgPassword,
		PostgresDB:       *pgDB,
		OutputFile:       *outputFile,
	}

	// Connect to PostgreSQL
	pgDBConn, err := ConnectPostgreSQL(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer CloseDB(pgDBConn)

	// Create GeoPackage
	gpkgDB, err := CreateGeoPackage(cfg.OutputFile)
	if err != nil {
		log.Fatalf("Failed to create GeoPackage: %v", err)
	}
	defer CloseDB(gpkgDB)

	// Initialize GeoPackage structure
	if err := InitGeoPackage(gpkgDB); err != nil {
		log.Fatalf("Failed to initialize GeoPackage: %v", err)
	}

	// Export data
	if err := ExportData(pgDBConn, gpkgDB); err != nil {
		log.Fatalf("Failed to export data: %v", err)
	}

	log.Printf("Successfully exported data to %s", cfg.OutputFile)
}
