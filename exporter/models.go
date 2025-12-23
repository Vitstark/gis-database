package main

import (
	"database/sql"
	"encoding/json"
)

// GeoJSONFeatureCollection represents a GeoJSON FeatureCollection
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

// GeoJSONFeature represents a GeoJSON Feature
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	ID         interface{}            `json:"id,omitempty"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// GeoJSONGeometry represents a GeoJSON Geometry
type GeoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	CRS         *GeoJSONCRS     `json:"crs,omitempty"`
}

// GeoJSONCRS represents a GeoJSON Coordinate Reference System
type GeoJSONCRS struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// CadastralObject represents a cadastral object from the database
type CadastralObject struct {
	Code                         int
	QuarterCode                  int
	LoadStatus                   string
	UpdateDate                   sql.NullTime
	Data                         string
	Area                         sql.NullInt64
	CostValue                    sql.NullFloat64
	PermittedUseEstablishedByDoc sql.NullString
	RightType                    sql.NullString
	Status                       sql.NullString
	LandRecordType               sql.NullString
	LandRecordSubtype            sql.NullString
	LandRecordCategoryType       sql.NullString
}
