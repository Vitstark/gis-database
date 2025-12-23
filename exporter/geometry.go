package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

// ConvertGeometryToGPKG converts a GeoJSON geometry to GPKG binary format
func ConvertGeometryToGPKG(geometry map[string]interface{}) ([]byte, error) {
	geomType, ok := geometry["type"].(string)
	if !ok {
		return nil, fmt.Errorf("geometry type not found")
	}

	coordinates, ok := geometry["coordinates"]
	if !ok {
		return nil, fmt.Errorf("coordinates not found")
	}

	// GPKG binary format:
	// Byte 0: Magic number (0x47 = 'G')
	// Byte 1: Version (0x00)
	// Byte 2: Flags
	// Bytes 3-6: SRS ID (little-endian int32)
	// Bytes 7-38: Envelope (optional, depending on flags)
	// Remaining: WKB geometry

	var wkb []byte
	var err error

	switch geomType {
	case "Polygon":
		wkb, err = polygonToWKB(coordinates)
	case "Point":
		wkb, err = pointToWKB(coordinates)
	case "LineString":
		wkb, err = lineStringToWKB(coordinates)
	case "MultiPolygon":
		wkb, err = multiPolygonToWKB(coordinates)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", geomType)
	}
	if err != nil {
		return nil, err
	}

	// Build GPKG binary header
	buf := make([]byte, 8)
	buf[0] = 0x47 // Magic number
	buf[1] = 0x00 // Version
	buf[2] = 0x01 // Flags: envelope indicator (standard envelope)

	// SRS ID: 3857 (little-endian) - bytes 3-6
	binary.LittleEndian.PutUint32(buf[3:7], 3857)
	buf[7] = 0x00 // Reserved byte (must be 0)

	// Envelope: min_x, max_x, min_y, max_y (8 bytes each = 32 bytes total)
	envelope := CalculateEnvelope(coordinates, geomType)
	envelopeBytes := make([]byte, 32)
	binary.LittleEndian.PutUint64(envelopeBytes[0:8], math.Float64bits(envelope[0]))   // min_x
	binary.LittleEndian.PutUint64(envelopeBytes[8:16], math.Float64bits(envelope[1]))  // max_x
	binary.LittleEndian.PutUint64(envelopeBytes[16:24], math.Float64bits(envelope[2])) // min_y
	binary.LittleEndian.PutUint64(envelopeBytes[24:32], math.Float64bits(envelope[3])) // max_y

	// Combine header + envelope + WKB
	result := append(buf, envelopeBytes...)
	result = append(result, wkb...)

	return result, nil
}

func polygonToWKB(coords interface{}) ([]byte, error) {
	coordsArray, ok := coords.([]interface{})
	if !ok || len(coordsArray) == 0 {
		return nil, fmt.Errorf("invalid polygon coordinates")
	}

	// Standard WKB format: byte order, type, num rings, rings
	// Note: SRID is in GPKG header, not in WKB
	buf := make([]byte, 9)                     // 1 byte order + 4 type + 4 num rings
	buf[0] = 1                                 // Little-endian
	binary.LittleEndian.PutUint32(buf[1:5], 3) // Polygon type (standard WKB, no SRID)

	rings := coordsArray
	binary.LittleEndian.PutUint32(buf[5:9], uint32(len(rings)))

	for _, ring := range rings {
		ringCoords, ok := ring.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid ring coordinates")
		}

		// Ring: num points + points
		ringBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(ringBuf, uint32(len(ringCoords)))

		for _, point := range ringCoords {
			pointCoords, ok := point.([]interface{})
			if !ok || len(pointCoords) < 2 {
				return nil, fmt.Errorf("invalid point coordinates")
			}

			x, err := toFloat64(pointCoords[0])
			if err != nil {
				return nil, fmt.Errorf("invalid x coordinate: %w", err)
			}
			y, err := toFloat64(pointCoords[1])
			if err != nil {
				return nil, fmt.Errorf("invalid y coordinate: %w", err)
			}

			pointBuf := make([]byte, 16) // 8 bytes x + 8 bytes y
			binary.LittleEndian.PutUint64(pointBuf[0:8], math.Float64bits(x))
			binary.LittleEndian.PutUint64(pointBuf[8:16], math.Float64bits(y))
			ringBuf = append(ringBuf, pointBuf...)
		}

		buf = append(buf, ringBuf...)
	}

	return buf, nil
}

func pointToWKB(coords interface{}) ([]byte, error) {
	coordsArray, ok := coords.([]interface{})
	if !ok || len(coordsArray) < 2 {
		return nil, fmt.Errorf("invalid point coordinates")
	}

	x, err := toFloat64(coordsArray[0])
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}
	y, err := toFloat64(coordsArray[1])
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	// Standard WKB format (SRID is in GPKG header)
	buf := make([]byte, 21)                    // 1 byte order + 4 type + 8 x + 8 y
	buf[0] = 1                                 // Little-endian
	binary.LittleEndian.PutUint32(buf[1:5], 1) // Point type (standard WKB)
	binary.LittleEndian.PutUint64(buf[5:13], math.Float64bits(x))
	binary.LittleEndian.PutUint64(buf[13:21], math.Float64bits(y))

	return buf, nil
}

func lineStringToWKB(coords interface{}) ([]byte, error) {
	coordsArray, ok := coords.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid linestring coordinates")
	}

	// Standard WKB format (SRID is in GPKG header)
	buf := make([]byte, 9)                     // 1 byte order + 4 type + 4 num points
	buf[0] = 1                                 // Little-endian
	binary.LittleEndian.PutUint32(buf[1:5], 2) // LineString type (standard WKB)
	binary.LittleEndian.PutUint32(buf[5:9], uint32(len(coordsArray)))

	for _, point := range coordsArray {
		pointCoords, ok := point.([]interface{})
		if !ok || len(pointCoords) < 2 {
			return nil, fmt.Errorf("invalid point coordinates")
		}

		x, err := toFloat64(pointCoords[0])
		if err != nil {
			return nil, fmt.Errorf("invalid x coordinate: %w", err)
		}
		y, err := toFloat64(pointCoords[1])
		if err != nil {
			return nil, fmt.Errorf("invalid y coordinate: %w", err)
		}

		pointBuf := make([]byte, 16)
		binary.LittleEndian.PutUint64(pointBuf[0:8], math.Float64bits(x))
		binary.LittleEndian.PutUint64(pointBuf[8:16], math.Float64bits(y))
		buf = append(buf, pointBuf...)
	}

	return buf, nil
}

func multiPolygonToWKB(coords interface{}) ([]byte, error) {
	coordsArray, ok := coords.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid multipolygon coordinates")
	}

	// Standard WKB format (SRID is in GPKG header)
	buf := make([]byte, 9)                     // 1 byte order + 4 type + 4 num polygons
	buf[0] = 1                                 // Little-endian
	binary.LittleEndian.PutUint32(buf[1:5], 6) // MultiPolygon type (standard WKB)
	binary.LittleEndian.PutUint32(buf[5:9], uint32(len(coordsArray)))

	for _, polygon := range coordsArray {
		polygonWKB, err := polygonToWKB(polygon)
		if err != nil {
			return nil, err
		}
		// Skip the first 9 bytes (byte order, type, num rings) and append the rest
		buf = append(buf, polygonWKB[9:]...)
	}

	return buf, nil
}

// CalculateEnvelope calculates the bounding box [min_x, max_x, min_y, max_y] for coordinates
func CalculateEnvelope(coords interface{}, geomType string) [4]float64 {
	// Returns [min_x, max_x, min_y, max_y]
	envelope := [4]float64{1e10, -1e10, 1e10, -1e10}

	var extractCoords func(interface{})
	extractCoords = func(c interface{}) {
		switch v := c.(type) {
		case []interface{}:
			if len(v) >= 2 {
				if x, err := toFloat64(v[0]); err == nil {
					if y, err := toFloat64(v[1]); err == nil {
						if x < envelope[0] {
							envelope[0] = x
						}
						if x > envelope[1] {
							envelope[1] = x
						}
						if y < envelope[2] {
							envelope[2] = y
						}
						if y > envelope[3] {
							envelope[3] = y
						}
					}
				}
			}
			for _, item := range v {
				extractCoords(item)
			}
		}
	}

	extractCoords(coords)
	return envelope
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
