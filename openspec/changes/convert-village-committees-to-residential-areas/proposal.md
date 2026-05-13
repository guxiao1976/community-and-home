## Why

The system currently lacks village-level residential area data. The `md_administrative_division` table contains 村委会 (village committee) records at level 5, which represent actual villages but are not accessible through the residential area query interface. Converting these records to `md_residential_area` entries will enable complete coverage of rural residential areas for data query and management.

## What Changes

- Extract all level=5 (村委会) records from `md_administrative_division` table
- Transform village committee names by removing "委会" suffix (e.g., "张家村委会" → "张家村")
- Resolve hierarchical IDs (street_id, county_id, city_id) by traversing parent_id relationships
- Insert transformed records into `md_residential_area` with community_type=2 (village)
- Implement deduplication logic to skip existing records with same name and county_id
- Generate conversion statistics (success count, skipped count, error count)

## Capabilities

### New Capabilities
- `village-data-conversion`: Batch conversion of administrative division village committee records to residential area entries with hierarchical ID resolution and name transformation

### Modified Capabilities
<!-- No existing capabilities are being modified -->

## Impact

- **Database**: New records inserted into `md_residential_area` table (estimated thousands of village records)
- **Data Model**: Utilizes existing `md_residential_area` schema, no schema changes required
- **Query Interface**: Village data becomes queryable through existing residential area query APIs
- **Data Source**: Introduces data_source=2 (administrative division import) as a new data origin marker
- **Dependencies**: Requires `md_administrative_division` table with complete parent_id relationships for level 2-5 records
