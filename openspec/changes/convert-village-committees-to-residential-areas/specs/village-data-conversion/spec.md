## ADDED Requirements

### Requirement: Extract village committee records
The system SHALL extract all records from `md_administrative_division` table where `level = 5`.

#### Scenario: Village committees found
- **WHEN** querying `md_administrative_division` with `level = 5`
- **THEN** system returns all village committee records with their id, code, name, and parent_id

#### Scenario: No village committees exist
- **WHEN** querying `md_administrative_division` with `level = 5` returns empty result
- **THEN** system completes with zero records processed

### Requirement: Resolve hierarchical administrative IDs
The system SHALL traverse parent_id relationships to resolve street_id (level 4), county_id (level 3), and city_id (level 2) for each village committee record.

#### Scenario: Complete hierarchy exists
- **WHEN** village committee has parent_id pointing to level 4 record, which has parent_id to level 3, which has parent_id to level 2
- **THEN** system resolves street_id from level 4, county_id from level 3, city_id from level 2

#### Scenario: Missing intermediate level
- **WHEN** village committee's parent chain skips a level (e.g., level 5 → level 3 directly)
- **THEN** system sets missing level ID to NULL and continues with available levels

#### Scenario: Broken parent chain
- **WHEN** parent_id references non-existent record
- **THEN** system logs error and skips this village committee record

### Requirement: Transform village name
The system SHALL transform village committee names by removing "委会" suffix while preserving the rest of the name.

#### Scenario: Standard village committee name
- **WHEN** source name is "张家村委会"
- **THEN** transformed name is "张家村"

#### Scenario: Name without 委会 suffix
- **WHEN** source name is "张家村"
- **THEN** transformed name remains "张家村"

#### Scenario: Name with 村委会 in middle
- **WHEN** source name is "村委会路社区"
- **THEN** transformed name is "村路社区" (only removes "委会")

### Requirement: Generate unique residential area code
The system SHALL generate a unique code for each residential area by appending a 4-digit sequence number to the village committee's administrative code.

#### Scenario: First residential area for village
- **WHEN** village committee code is "370502001201" and no existing residential areas use this prefix
- **THEN** generated code is "3705020012010001"

#### Scenario: Existing residential areas with same prefix
- **WHEN** village committee code is "370502001201" and codes "3705020012010001" through "3705020012010003" exist
- **THEN** generated code is "3705020012010004"

### Requirement: Check for duplicate records
The system SHALL check if a residential area with the same name and county_id already exists before insertion.

#### Scenario: No duplicate exists
- **WHEN** no record in `md_residential_area` has matching name and county_id
- **THEN** system proceeds with insertion

#### Scenario: Duplicate exists
- **WHEN** a record in `md_residential_area` has matching name and county_id
- **THEN** system skips insertion and increments skipped count

### Requirement: Insert residential area record
The system SHALL insert transformed village data into `md_residential_area` table with specified field mappings.

#### Scenario: Successful insertion
- **WHEN** all validations pass and no duplicate exists
- **THEN** system inserts record with:
  - name: transformed village name
  - code: generated unique code
  - county_id: resolved from level 3
  - city_id: resolved from level 2
  - street_id: resolved from level 4
  - community_type: 2 (village)
  - submission_status: 2 (approved)
  - data_source: 2 (administrative division import)
  - created_at: current timestamp
  - updated_at: current timestamp

#### Scenario: Database constraint violation
- **WHEN** insertion fails due to database constraint (e.g., foreign key violation)
- **THEN** system logs error with village name and constraint details, increments error count

### Requirement: Generate conversion statistics
The system SHALL track and report conversion statistics including success count, skipped count, and error count.

#### Scenario: Conversion completes
- **WHEN** all village committee records have been processed
- **THEN** system outputs summary with:
  - Total records processed
  - Successfully inserted count
  - Skipped (duplicate) count
  - Error count with sample error messages

#### Scenario: Conversion interrupted
- **WHEN** conversion process is interrupted (e.g., database connection lost)
- **THEN** system outputs partial statistics up to interruption point
