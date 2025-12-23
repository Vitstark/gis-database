#!/bin/bash

# Script to generate GeoJSON exports with all possible groupings
# Usage: ./generate_all_groupings.sh [output_directory]

set -e

# Default output directory
OUTPUT_DIR="${1:-geojson_exports}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Generating GeoJSON exports with all groupings ===${NC}"
echo "Output directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if exporter exists
if [ ! -f "./export_geojson" ]; then
    echo -e "${YELLOW}Building exporter...${NC}"
    go build -o export_geojson export_geojson.go
fi

# Function to run exporter and check result
run_export() {
    local group_by="$1"
    local output_name="$2"
    local description="$3"
    
    echo -e "${GREEN}Generating: $description${NC}"
    echo "  Group by: $group_by"
    echo "  Output: $OUTPUT_DIR/$output_name"
    
    if ./export_geojson -group-by "$group_by" -output "$OUTPUT_DIR/$output_name" 2>&1 | grep -E "(Total exported|Created:|Failed)"; then
        echo -e "${GREEN}  ✓ Success${NC}"
    else
        echo -e "${YELLOW}  ⚠ Check output above${NC}"
    fi
    echo ""
}

# Generate single file (no grouping) for reference
echo -e "${BLUE}=== Single File (No Grouping) ===${NC}"
echo "Generating: Single FeatureCollection with all features"
echo "  Output: $OUTPUT_DIR/cadastral_all.geojson"
if ./export_geojson -output "$OUTPUT_DIR/cadastral_all.geojson" 2>&1 | grep -E "(Total exported|Failed)"; then
    echo -e "${GREEN}  ✓ Success${NC}"
else
    echo -e "${YELLOW}  ⚠ Check output above${NC}"
fi
echo ""

# Generate grouped exports
echo -e "${BLUE}=== Grouped Exports ===${NC}"

# Group by status
run_export "status" "cadastral_by_status" "Grouped by status (Ранее учтенный, Учтенный, etc.)"

# Group by quarter_code
run_export "quarter_code" "cadastral_by_quarter" "Grouped by quarter code"

# Group by right_type
run_export "right_type" "cadastral_by_right_type" "Grouped by right type (Собственность, etc.)"

# Group by land_record_type
run_export "land_record_type" "cadastral_by_land_record_type" "Grouped by land record type"

# Group by land_record_subtype
run_export "land_record_subtype" "cadastral_by_land_record_subtype" "Grouped by land record subtype"

# Group by land_record_category_type
run_export "land_record_category_type" "cadastral_by_land_record_category" "Grouped by land record category"

# Group by load_status
run_export "load_status" "cadastral_by_load_status" "Grouped by load status"

# Summary
echo -e "${BLUE}=== Summary ===${NC}"
echo "All exports completed!"
echo ""
echo "Generated files and directories:"
ls -lh "$OUTPUT_DIR" | tail -n +2 | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo -e "${GREEN}Total size:${NC}"
du -sh "$OUTPUT_DIR"
echo ""
echo -e "${BLUE}Directory structure:${NC}"
find "$OUTPUT_DIR" -type f -name "*.geojson" | sort | sed 's|^|  |'
echo ""
echo -e "${GREEN}✓ All exports generated in: $OUTPUT_DIR${NC}"

