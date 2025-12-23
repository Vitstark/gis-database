package main

import (
	"database/sql"
	"fmt"
)

// InitGeoPackage initializes a GeoPackage file with required metadata tables
func InitGeoPackage(db *sql.DB) error {
	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create gpkg_spatial_ref_sys table
	if err := createSpatialRefSysTable(db); err != nil {
		return err
	}

	// Insert required SRS entries
	if err := insertSpatialRefSystems(db); err != nil {
		return err
	}

	// Create gpkg_contents table
	if err := createContentsTable(db); err != nil {
		return err
	}

	// Create gpkg_geometry_columns table
	if err := createGeometryColumnsTable(db); err != nil {
		return err
	}

	// Create cadastral_objects table
	if err := createCadastralObjectsTable(db); err != nil {
		return err
	}

	// Register table in gpkg_contents
	if err := registerTableInContents(db); err != nil {
		return err
	}

	// Register geometry column
	if err := registerGeometryColumn(db); err != nil {
		return err
	}

	return nil
}

func createSpatialRefSysTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gpkg_spatial_ref_sys (
			srs_name TEXT NOT NULL,
			srs_id INTEGER NOT NULL PRIMARY KEY,
			organization TEXT NOT NULL,
			organization_coordsys_id INTEGER NOT NULL,
			definition TEXT NOT NULL,
			description TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gpkg_spatial_ref_sys: %w", err)
	}
	return nil
}

func insertSpatialRefSystems(db *sql.DB) error {
	// Insert EPSG:3857 (Web Mercator) - the CRS used in the data
	_, err := db.Exec(`
		INSERT OR REPLACE INTO gpkg_spatial_ref_sys 
		(srs_name, srs_id, organization, organization_coordsys_id, definition, description)
		VALUES 
		('WGS 84 / Pseudo-Mercator', 3857, 'EPSG', 3857, 
		 'PROJCS["WGS 84 / Pseudo-Mercator",GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]],PROJECTION["Mercator_1SP"],PARAMETER["central_meridian",0],PARAMETER["scale_factor",1],PARAMETER["false_easting",0],PARAMETER["false_northing",0],UNIT["metre",1,AUTHORITY["EPSG","9001"]],AXIS["X",EAST],AXIS["Y",NORTH],EXTENSION["PROJ4","+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs"],AUTHORITY["EPSG","3857"]]',
		 'Popular Visualisation CRS / Mercator')
	`)
	if err != nil {
		return fmt.Errorf("failed to insert EPSG:3857: %w", err)
	}

	// Insert EPSG:4326 (WGS84) - required by GeoPackage spec
	_, err = db.Exec(`
		INSERT OR REPLACE INTO gpkg_spatial_ref_sys 
		(srs_name, srs_id, organization, organization_coordsys_id, definition, description)
		VALUES 
		('WGS 84', 4326, 'EPSG', 4326,
		 'GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]',
		 'WGS 84')
	`)
	if err != nil {
		return fmt.Errorf("failed to insert EPSG:4326: %w", err)
	}

	return nil
}

func createContentsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gpkg_contents (
			table_name TEXT NOT NULL PRIMARY KEY,
			data_type TEXT NOT NULL,
			identifier TEXT UNIQUE,
			description TEXT,
			last_change DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			min_x DOUBLE,
			min_y DOUBLE,
			max_x DOUBLE,
			max_y DOUBLE,
			srs_id INTEGER,
			CONSTRAINT fk_gc_srs FOREIGN KEY (srs_id) REFERENCES gpkg_spatial_ref_sys(srs_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gpkg_contents: %w", err)
	}
	return nil
}

func createGeometryColumnsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS gpkg_geometry_columns (
			table_name TEXT NOT NULL,
			column_name TEXT NOT NULL,
			geometry_type_name TEXT NOT NULL,
			srs_id INTEGER NOT NULL,
			z TINYINT NOT NULL,
			m TINYINT NOT NULL,
			CONSTRAINT pk_geom_cols PRIMARY KEY (table_name, column_name),
			CONSTRAINT fk_gc_tn FOREIGN KEY (table_name) REFERENCES gpkg_contents(table_name),
			CONSTRAINT fk_gc_srs FOREIGN KEY (srs_id) REFERENCES gpkg_spatial_ref_sys(srs_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gpkg_geometry_columns: %w", err)
	}
	return nil
}

func createCadastralObjectsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cadastral_objects (
			code INTEGER NOT NULL PRIMARY KEY,
			quarter_code INTEGER NOT NULL,
			load_status TEXT,
			update_date DATE,
			area INTEGER,
			cost_value REAL,
			permitted_use_established_by_document TEXT,
			right_type TEXT,
			status TEXT,
			land_record_type TEXT,
			land_record_subtype TEXT,
			land_record_category_type TEXT,
			geometry BLOB NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create cadastral_objects table: %w", err)
	}
	return nil
}

func registerTableInContents(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO gpkg_contents 
		(table_name, data_type, identifier, description, srs_id)
		VALUES 
		('cadastral_objects', 'features', 'Cadastral Objects', 'Cadastral objects from Kazan', 3857)
	`)
	if err != nil {
		return fmt.Errorf("failed to register table in gpkg_contents: %w", err)
	}
	return nil
}

func registerGeometryColumn(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO gpkg_geometry_columns 
		(table_name, column_name, geometry_type_name, srs_id, z, m)
		VALUES 
		('cadastral_objects', 'geometry', 'POLYGON', 3857, 0, 0)
	`)
	if err != nil {
		return fmt.Errorf("failed to register geometry column: %w", err)
	}
	return nil
}
