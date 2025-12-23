package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		pgHost     = flag.String("pg-host", "localhost", "PostgreSQL host")
		pgPort     = flag.Int("pg-port", 5432, "PostgreSQL port")
		pgUser     = flag.String("pg-user", "postgres", "PostgreSQL user")
		pgPassword = flag.String("pg-password", "postgres", "PostgreSQL password")
		pgDB       = flag.String("pg-db", "postgres", "PostgreSQL database name")
		outputFile = flag.String("output", "cadastral.geojson", "Output GeoJSON file path")
		groupBy    = flag.String("group-by", "", "Group features by property (e.g., 'quarter_code', 'area_code', 'status'). Creates multiple FeatureCollections, one per unique value")
	)
	flag.Parse()

	// Connect to PostgreSQL
	pgDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*pgHost, *pgPort, *pgUser, *pgPassword, *pgDB)

	pgDBConn, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDBConn.Close()

	if err := pgDBConn.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	// Export to GeoJSON
	if err := exportToGeoJSON(pgDBConn, *outputFile, *groupBy); err != nil {
		log.Fatalf("Failed to export data: %v", err)
	}

	log.Printf("Successfully exported data to %s", *outputFile)
}

func exportToGeoJSON(pgDB *sql.DB, outputFile string, groupByProperty string) error {
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

	// If grouping is enabled, use map to group by property value
	// Otherwise, use single FeatureCollection
	var groupedFeatures map[string][]interface{}
	var singleFeatures []interface{}
	var count int

	if groupByProperty != "" {
		groupedFeatures = make(map[string][]interface{})
		log.Printf("Grouping features by property: %s", groupByProperty)
	} else {
		singleFeatures = []interface{}{}
	}

	for rows.Next() {
		var code, quarterCode int
		var loadStatus, dataStr string
		var updateDate sql.NullTime
		var area sql.NullInt64
		var costValue sql.NullFloat64
		var permittedUse, rightType, status sql.NullString
		var landRecordType, landRecordSubtype, landRecordCategoryType sql.NullString

		if err := rows.Scan(
			&code,
			&quarterCode,
			&loadStatus,
			&updateDate,
			&dataStr,
			&area,
			&costValue,
			&permittedUse,
			&rightType,
			&status,
			&landRecordType,
			&landRecordSubtype,
			&landRecordCategoryType,
		); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Parse the JSON data
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			log.Printf("Failed to parse JSON for object %d: %v", code, err)
			continue
		}

		// Extract FeatureCollection from data
		dataObj, ok := data["data"].(map[string]interface{})
		if !ok {
			log.Printf("No 'data' field in JSON for object %d", code)
			continue
		}

		dataFeatures, ok := dataObj["features"].([]interface{})
		if !ok || len(dataFeatures) == 0 {
			log.Printf("No features in JSON for object %d", code)
			continue
		}

		// Get first feature and update its properties with database fields
		feature, ok := dataFeatures[0].(map[string]interface{})
		if !ok {
			log.Printf("Invalid feature structure for object %d", code)
			continue
		}

		// Convert geometry coordinates from EPSG:3857 to EPSG:4326 (WGS84)
		// GeoJSON standard requires WGS84 coordinates
		if err := convertGeometryToWGS84(feature["geometry"]); err != nil {
			log.Printf("Failed to convert geometry for object %d: %v", code, err)
			continue
		}

		// Update properties with database fields
		properties := map[string]interface{}{
			"code":                                  code,
			"quarter_code":                          quarterCode,
			"load_status":                           loadStatus,
			"area":                                  getNullableInt64(area),
			"cost_value":                            getNullableFloat64(costValue),
			"permitted_use_established_by_document": getNullableString(permittedUse),
			"right_type":                            getNullableString(rightType),
			"status":                                getNullableString(status),
			"land_record_type":                      getNullableString(landRecordType),
			"land_record_subtype":                   getNullableString(landRecordSubtype),
			"land_record_category_type":             getNullableString(landRecordCategoryType),
		}

		if updateDate.Valid {
			properties["update_date"] = updateDate.Time.Format("2006-01-02")
		}

		// Merge with existing properties if any
		if existingProps, ok := feature["properties"].(map[string]interface{}); ok {
			for k, v := range existingProps {
				if _, exists := properties[k]; !exists {
					properties[k] = v
				}
			}
		}

		feature["properties"] = properties

		// Group by property if specified
		if groupByProperty != "" {
			groupValue := getGroupValue(properties, groupByProperty)
			if groupValue == "" {
				groupValue = "unknown"
			}
			groupedFeatures[groupValue] = append(groupedFeatures[groupValue], feature)
		} else {
			singleFeatures = append(singleFeatures, feature)
		}

		count++
		if count%100 == 0 {
			log.Printf("Processed %d objects...", count)
		}
	}

	// Write to file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if groupByProperty != "" {
		// Create directory for multiple files
		outputDir := outputFile
		// If output looks like a file (has extension), extract directory
		if strings.Contains(outputDir, ".") && filepath.Ext(outputDir) != "" {
			outputDir = filepath.Dir(outputDir)
			if outputDir == "." {
				outputDir = filepath.Base(outputFile)
				outputDir = strings.TrimSuffix(outputDir, filepath.Ext(outputDir))
			}
		}

		// Remove existing file/directory if it exists
		if info, err := os.Stat(outputDir); err == nil {
			if !info.IsDir() {
				os.Remove(outputDir)
			}
		}

		// Ensure directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Create one file per group
		baseName := "cadastral"
		if strings.Contains(outputFile, ".") {
			ext := filepath.Ext(outputFile)
			baseName = strings.TrimSuffix(filepath.Base(outputFile), ext)
		}

		for groupValue, features := range groupedFeatures {
			// Sanitize group value for filename
			safeGroupValue := sanitizeFilename(groupValue)
			// Include the grouping field name in the filename
			filename := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%s.geojson", baseName, groupByProperty, safeGroupValue))

			// Create FeatureCollection for this group
			featureCollection := map[string]interface{}{
				"type":     "FeatureCollection",
				"features": features,
			}

			// Write to file
			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", filename, err)
			}

			fileEncoder := json.NewEncoder(file)
			fileEncoder.SetIndent("", "  ")
			if err := fileEncoder.Encode(featureCollection); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode GeoJSON for %s: %w", filename, err)
			}
			file.Close()

			log.Printf("  Created: %s (%d features)", filename, len(features))
		}
		log.Printf("Total exported: %d features in %d files in directory: %s", count, len(groupedFeatures), outputDir)
	} else {
		// Single FeatureCollection
		featureCollection := map[string]interface{}{
			"type":     "FeatureCollection",
			"features": singleFeatures,
		}
		if err := encoder.Encode(featureCollection); err != nil {
			return fmt.Errorf("failed to encode GeoJSON: %w", err)
		}
		log.Printf("Total exported: %d features", count)
	}

	return nil
}

// getGroupValue extracts the grouping value from properties
func getGroupValue(properties map[string]interface{}, propertyName string) string {
	value, ok := properties[propertyName]
	if !ok {
		return ""
	}

	// Convert to string
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Remove any remaining problematic characters
	result = strings.TrimSpace(result)
	if result == "" {
		result = "unknown"
	}
	return result
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

// convertGeometryToWGS84 converts geometry coordinates from EPSG:3857 (Web Mercator) to EPSG:4326 (WGS84)
func convertGeometryToWGS84(geometry interface{}) error {
	geom, ok := geometry.(map[string]interface{})
	if !ok {
		return fmt.Errorf("geometry is not a map")
	}

	coords, ok := geom["coordinates"]
	if !ok {
		return fmt.Errorf("no coordinates in geometry")
	}

	// Convert coordinates based on geometry type
	geomType, _ := geom["type"].(string)
	switch geomType {
	case "Point":
		coordsArray, ok := coords.([]interface{})
		if !ok || len(coordsArray) < 2 {
			return fmt.Errorf("invalid point coordinates")
		}
		x, _ := toFloat64(coordsArray[0])
		y, _ := toFloat64(coordsArray[1])
		lon, lat := webMercatorToWGS84(x, y)
		geom["coordinates"] = []float64{lon, lat}

	case "LineString":
		coordsArray, ok := coords.([]interface{})
		if !ok {
			return fmt.Errorf("invalid linestring coordinates")
		}
		converted := make([][]float64, len(coordsArray))
		for i, point := range coordsArray {
			pointArray, ok := point.([]interface{})
			if !ok || len(pointArray) < 2 {
				return fmt.Errorf("invalid point in linestring")
			}
			x, _ := toFloat64(pointArray[0])
			y, _ := toFloat64(pointArray[1])
			lon, lat := webMercatorToWGS84(x, y)
			converted[i] = []float64{lon, lat}
		}
		geom["coordinates"] = converted

	case "Polygon":
		ringsArray, ok := coords.([]interface{})
		if !ok {
			return fmt.Errorf("invalid polygon coordinates")
		}
		converted := make([][][]float64, len(ringsArray))
		for i, ring := range ringsArray {
			ringArray, ok := ring.([]interface{})
			if !ok {
				return fmt.Errorf("invalid ring in polygon")
			}
			convertedRing := make([][]float64, len(ringArray))
			for j, point := range ringArray {
				pointArray, ok := point.([]interface{})
				if !ok || len(pointArray) < 2 {
					return fmt.Errorf("invalid point in ring")
				}
				x, _ := toFloat64(pointArray[0])
				y, _ := toFloat64(pointArray[1])
				lon, lat := webMercatorToWGS84(x, y)
				convertedRing[j] = []float64{lon, lat}
			}
			converted[i] = convertedRing
		}
		geom["coordinates"] = converted

	case "MultiPolygon":
		multiPolyArray, ok := coords.([]interface{})
		if !ok {
			return fmt.Errorf("invalid multipolygon coordinates")
		}
		converted := make([][][][]float64, len(multiPolyArray))
		for i, polygon := range multiPolyArray {
			ringsArray, ok := polygon.([]interface{})
			if !ok {
				return fmt.Errorf("invalid polygon in multipolygon")
			}
			convertedPoly := make([][][]float64, len(ringsArray))
			for j, ring := range ringsArray {
				ringArray, ok := ring.([]interface{})
				if !ok {
					return fmt.Errorf("invalid ring in multipolygon")
				}
				convertedRing := make([][]float64, len(ringArray))
				for k, point := range ringArray {
					pointArray, ok := point.([]interface{})
					if !ok || len(pointArray) < 2 {
						return fmt.Errorf("invalid point in multipolygon")
					}
					x, _ := toFloat64(pointArray[0])
					y, _ := toFloat64(pointArray[1])
					lon, lat := webMercatorToWGS84(x, y)
					convertedRing[k] = []float64{lon, lat}
				}
				convertedPoly[j] = convertedRing
			}
			converted[i] = convertedPoly
		}
		geom["coordinates"] = converted

	default:
		return fmt.Errorf("unsupported geometry type: %s", geomType)
	}

	// Remove CRS field if present (GeoJSON uses WGS84 by default)
	delete(geom, "crs")
	return nil
}

// webMercatorToWGS84 converts Web Mercator (EPSG:3857) coordinates to WGS84 (EPSG:4326)
func webMercatorToWGS84(x, y float64) (lon, lat float64) {
	// Web Mercator to WGS84 conversion
	lon = x / 20037508.34 * 180.0
	lat = y / 20037508.34 * 180.0
	lat = 180.0 / math.Pi * (2.0*math.Atan(math.Exp(lat*math.Pi/180.0)) - math.Pi/2.0)
	return lon, lat
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case json.Number:
		return val.Float64()
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
