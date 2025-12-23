package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

// ExportData exports cadastral objects from PostgreSQL to GeoPackage
func ExportData(pgDB *sql.DB, gpkgDB *sql.DB) error {
	// Query cadastral objects with their data
	rows, err := pgDB.Query(`
		SELECT 
			o.code,
			o.quarter_code,
			o.load_status::text,
			o.update_date,
			o.data::text,
			o.area,
			o.cost_value,
			o.permitted_use_established_by_document,
			o.right_type,
			o.status,
			o.land_record_type,
			o.land_record_subtype,
			o.land_record_category_type
		FROM object o
		WHERE o.data IS NOT NULL
		AND o.load_status = 'SUCCESS'
	`)
	if err != nil {
		return fmt.Errorf("failed to query objects: %w", err)
	}
	defer rows.Close()

	// Prepare insert statement
	stmt, err := gpkgDB.Prepare(`
		INSERT INTO cadastral_objects 
		(code, quarter_code, load_status, update_date, area, cost_value,
		 permitted_use_established_by_document, right_type, status,
		 land_record_type, land_record_subtype, land_record_category_type, geometry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	var count int
	var globalMinX, globalMinY, globalMaxX, globalMaxY float64
	globalMinX = 1e10
	globalMinY = 1e10
	globalMaxX = -1e10
	globalMaxY = -1e10

	for rows.Next() {
		var obj CadastralObject
		if err := rows.Scan(
			&obj.Code,
			&obj.QuarterCode,
			&obj.LoadStatus,
			&obj.UpdateDate,
			&obj.Data,
			&obj.Area,
			&obj.CostValue,
			&obj.PermittedUseEstablishedByDoc,
			&obj.RightType,
			&obj.Status,
			&obj.LandRecordType,
			&obj.LandRecordSubtype,
			&obj.LandRecordCategoryType,
		); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Extract geometry from GeoJSON
		geometry, err := extractGeometryFromJSON(obj.Data)
		if err != nil {
			log.Printf("Failed to extract geometry for object %d: %v", obj.Code, err)
			continue
		}

		// Calculate envelope for this geometry
		coordinates, ok := geometry["coordinates"]
		if ok {
			geomType, _ := geometry["type"].(string)
			envelope := CalculateEnvelope(coordinates, geomType)
			if envelope[0] < globalMinX {
				globalMinX = envelope[0]
			}
			if envelope[1] > globalMaxX {
				globalMaxX = envelope[1]
			}
			if envelope[2] < globalMinY {
				globalMinY = envelope[2]
			}
			if envelope[3] > globalMaxY {
				globalMaxY = envelope[3]
			}
		}

		// Convert geometry to GPKG binary format
		gpkgGeometry, err := ConvertGeometryToGPKG(geometry)
		if err != nil {
			log.Printf("Failed to convert geometry for object %d: %v", obj.Code, err)
			continue
		}

		// Insert into GeoPackage
		var updateDate interface{}
		if obj.UpdateDate.Valid {
			updateDate = obj.UpdateDate.Time.Format("2006-01-02")
		}

		_, err = stmt.Exec(
			obj.Code,
			obj.QuarterCode,
			obj.LoadStatus,
			updateDate,
			getNullableInt64(obj.Area),
			getNullableFloat64(obj.CostValue),
			getNullableString(obj.PermittedUseEstablishedByDoc),
			getNullableString(obj.RightType),
			getNullableString(obj.Status),
			getNullableString(obj.LandRecordType),
			getNullableString(obj.LandRecordSubtype),
			getNullableString(obj.LandRecordCategoryType),
			gpkgGeometry,
		)
		if err != nil {
			log.Printf("Failed to insert object %d: %v", obj.Code, err)
			continue
		}

		count++
		if count%100 == 0 {
			log.Printf("Exported %d objects...", count)
		}
	}

	log.Printf("Total exported: %d objects", count)

	// Update envelope in gpkg_contents with calculated bounds
	if count > 0 {
		if err := updateContentsEnvelope(gpkgDB, globalMinX, globalMinY, globalMaxX, globalMaxY); err != nil {
			return fmt.Errorf("failed to update envelope: %w", err)
		}
	}

	return nil
}

// updateContentsEnvelope updates the envelope in gpkg_contents
func updateContentsEnvelope(db *sql.DB, minX, minY, maxX, maxY float64) error {
	_, err := db.Exec(`
		UPDATE gpkg_contents 
		SET min_x = ?, min_y = ?, max_x = ?, max_y = ?
		WHERE table_name = 'cadastral_objects'
	`, minX, minY, maxX, maxY)

	return err
}

// extractGeometryFromJSON extracts the geometry from a GeoJSON data string
func extractGeometryFromJSON(dataStr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract FeatureCollection
	dataObj, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no 'data' field in JSON")
	}

	features, ok := dataObj["features"].([]interface{})
	if !ok || len(features) == 0 {
		return nil, fmt.Errorf("no features in JSON")
	}

	// Get first feature's geometry
	feature, ok := features[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid feature structure")
	}

	geometry, ok := feature["geometry"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no geometry in feature")
	}

	return geometry, nil
}

// Helper functions for nullable SQL types
func getNullableInt64(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}

func getNullableFloat64(n sql.NullFloat64) interface{} {
	if n.Valid {
		return n.Float64
	}
	return nil
}

func getNullableString(n sql.NullString) interface{} {
	if n.Valid {
		return n.String
	}
	return nil
}
