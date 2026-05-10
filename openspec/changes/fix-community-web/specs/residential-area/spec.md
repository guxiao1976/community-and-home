## ADDED Requirements

### Requirement: SQL schema MUST define md_residential_area table with administrative division fields
The `masterdata_schema.sql` script SHALL define the table `md_residential_area` (replacing the old `md_community` table). The table MUST include the following administrative division fields: `county_id` (level 3), `street_id` (level 4), `community_div_id` (level 5), and `code` (unique residential area code). Each division field SHALL reference `md_administrative_division(id)` via foreign key.

#### Scenario: Schema file creates md_residential_area table
- **WHEN** `masterdata_schema.sql` is executed against a fresh database
- **THEN** table `md_residential_area` is created with columns: `id`, `county_id`, `street_id`, `community_div_id`, `code`, `name`, `address`, `area`, `population`, `community_type`, `submission_status`, `submitter_id`, `submit_time`, `reviewer_id`, `review_time`, `review_notes`, `created_time`, `updated_time`, `delete_time`
- **AND** foreign keys `fk_ra_county`, `fk_ra_street`, `fk_ra_community_div` are created referencing `md_administrative_division(id)`
- **AND** table `md_community` does NOT exist

#### Scenario: Old md_community table is dropped
- **WHEN** `masterdata_schema.sql` is executed
- **THEN** the script contains `DROP TABLE IF EXISTS md_community` before creating `md_residential_area`
- **AND** the old `fk_community_division` constraint no longer exists

### Requirement: Seed data MUST populate md_residential_area with new field values
The `masterdata_seed.sql` script SHALL insert sample data into `md_residential_area` (not `md_community`). Each record MUST include `county_id`, `street_id`, `community_div_id`, and `code` values derived from the administrative division hierarchy.

#### Scenario: Seed data inserts into md_residential_area
- **WHEN** `masterdata_seed.sql` is executed after `masterdata_schema.sql`
- **THEN** 5 sample residential area records are inserted into `md_residential_area`
- **AND** each record has non-null `county_id`, `street_id`, `community_div_id`, and `code` values
- **AND** no data is inserted into `md_community`

#### Scenario: Seed county_id and street_id match administrative division hierarchy
- **WHEN** a seed record has `community_div_id = 11411` (中关村东路社区, under 中关村街道, under 海淀区)
- **THEN** its `county_id` SHALL be `114` (海淀区) and `street_id` SHALL be `1141` (中关村街道)

### Requirement: Documentation MUST reference md_residential_area instead of md_community
All project documentation files SHALL replace references to `md_community` with `md_residential_area`. The relationship diagram in `data-model.md` SHALL show `md_residential_area` connected to `md_administrative_division` via three foreign keys (county, street, community_div) instead of a single `division_id`.

#### Scenario: PROJECT_STRUCTURE.md uses new table name
- **WHEN** `PROJECT_STRUCTURE.md` is reviewed
- **THEN** it references `md_residential_area (住宅小区表)` instead of `md_community (社区表)`

#### Scenario: data-model.md shows updated table relationships
- **WHEN** `specs/001-identity-masterdata/data-model.md` is reviewed
- **THEN** the table section header reads `### md_residential_area` instead of `### md_community`
- **AND** the relationship diagram shows three foreign keys: `county_id`, `street_id`, `community_div_id` referencing `md_administrative_division`

#### Scenario: tasks.md references new model generation task
- **WHEN** `specs/001-identity-masterdata/tasks.md` is reviewed
- **THEN** it references `mdResidentialAreaModel` instead of `mdcommunitymodel`

#### Scenario: Audit log entity_type uses new table name
- **WHEN** `specs/001-identity-masterdata/contracts/masterdata-api.md` is reviewed
- **THEN** the audit log entity_type example shows `"md_residential_area"` instead of `"md_community"`
