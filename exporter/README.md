# Cadastral Data Exporter

Go applications that export cadastral objects from a PostgreSQL database to GIS formats:
- **GeoPackage** (`.gpkg`) - Standardized SQLite-based format
- **GeoJSON** (`.geojson`) - Simple JSON-based format, directly importable in QGIS

## Requirements

- Go 1.21 or later
- PostgreSQL database with cadastral data
- SQLite3 (for GeoPackage support)

## Installation

```bash
cd exporter
go mod download
```

## Usage

### Export to GeoPackage

```bash
go run main.go [flags]
# or
go build -o exporter . && ./exporter [flags]
```

### Export to GeoJSON

```bash
go run export_geojson.go [flags]
# or
go build -o export_geojson export_geojson.go && ./export_geojson [flags]
```

### Flags (both exporters)

- `-pg-host`: PostgreSQL host (default: "localhost")
- `-pg-port`: PostgreSQL port (default: 5432)
- `-pg-user`: PostgreSQL user (default: "postgres")
- `-pg-password`: PostgreSQL password (default: "postgres")
- `-pg-db`: PostgreSQL database name (default: "postgres")
- `-output`: Output file path
  - GeoPackage: default `"cadastral.gpkg"`
  - GeoJSON: default `"cadastral.geojson"`
- `-group-by`: (GeoJSON only) Group features by property value, creating multiple files in a directory. 
  - Example: `-group-by quarter_code` creates one file per unique quarter_code
  - Example: `-group-by status` creates one file per unique status value
  - If not specified, creates a single FeatureCollection file (standard GeoJSON)
  - When used, the `-output` parameter specifies the directory name (or file path, from which directory is derived)

### Examples

**Export to GeoPackage:**
```bash
go run main.go \
  -pg-host localhost \
  -pg-port 5432 \
  -pg-user postgres \
  -pg-password postgres \
  -pg-db postgres \
  -output kazan_cadastral.gpkg
```

**Export to GeoJSON (single layer):**
```bash
go run export_geojson.go \
  -pg-host localhost \
  -pg-port 5432 \
  -pg-user postgres \
  -pg-password postgres \
  -pg-db postgres \
  -output kazan_cadastral.geojson
```

**Export to GeoJSON (multiple layers by quarter_code):**
```bash
go run export_geojson.go \
  -group-by quarter_code \
  -output kazan_cadastral_by_quarter.geojson
```

**Export to GeoJSON (multiple layers by status):**
```bash
go run export_geojson.go \
  -group-by status \
  -output kazan_cadastral_by_status.geojson
```

## Output Formats

### GeoPackage (`.gpkg`)

Creates a standardized GeoPackage file containing:

- **Table**: `cadastral_objects`
- **Geometry**: Polygons in EPSG:3857 (Web Mercator)
- **Attributes**: All cadastral object fields including code, area, cost_value, status, etc.

The GeoPackage file can be opened in GIS software such as QGIS, ArcGIS, or any other tool that supports the GeoPackage format.

### GeoJSON (`.geojson`)

Creates a single GeoJSON FeatureCollection file containing:

- **Format**: Standard GeoJSON FeatureCollection
- **Geometry**: Polygons in EPSG:3857 (Web Mercator)
- **Attributes**: All cadastral object fields merged with original GeoJSON properties

**Advantages of GeoJSON:**
- ✅ Directly importable in QGIS (drag and drop)
- ✅ Human-readable format
- ✅ Easy to share and view in web browsers
- ✅ Smaller file size for simple use cases

**To import in QGIS:**
1. Open QGIS
2. Drag and drop the `.geojson` file into QGIS, or
3. Go to `Layer` → `Add Layer` → `Add Vector Layer...` and select the file

**Multiple Files (with `-group-by`):**
- Creates a directory containing multiple GeoJSON files, one per unique value of the selected property
- Each file is a standard single FeatureCollection that QGIS can import directly
- Files are named: `{basename}_{fieldname}_{groupvalue}.geojson`
- Example: `-group-by status -output cadastral_by_status` creates:
  - `cadastral_by_status/cadastral_status_Ранее_учтенный.geojson`
  - `cadastral_by_status/cadastral_status_Учтенный.geojson`
- The field name in the filename makes it clear which property was used for grouping
- Each file can be imported separately in QGIS as its own layer

## Database Schema

The application expects the following PostgreSQL schema:

- `object` table with:
  - `code` (integer)
  - `quarter_code` (integer)
  - `load_status` (enum: 'NEW', 'SUCCESS', 'ERROR', 'NOT FOUND')
  - `update_date` (date)
  - `data` (jsonb) - containing GeoJSON FeatureCollection
  - Additional attribute fields

Only objects with `load_status = 'SUCCESS'` and non-null `data` are exported.

