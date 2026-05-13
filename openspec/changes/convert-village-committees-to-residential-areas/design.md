## Context

The system currently has two separate data sources for residential areas:
1. **High-level data**: Cities and counties with residential areas imported from Amap POI API
2. **Village-level data**: Village committees (村委会) stored in `md_administrative_division` table at level 5

The administrative division table contains complete hierarchical relationships (province → city → county → street → village), but village data is not accessible through the residential area query interface. This design addresses the conversion of village committee records to residential area entries.

**Current State:**
- `md_administrative_division` table has level 5 records representing village committees
- `md_residential_area` table has community_type field with values: 1 (residential complex), 2 (village), 3 (mixed)
- Existing residential areas primarily come from Amap sync (data_source=1) and manual entry (data_source=0)

**Constraints:**
- Must preserve existing residential area data
- Must avoid duplicate entries (same name + county_id)
- Must maintain referential integrity with administrative division hierarchy
- Conversion should be idempotent (can be run multiple times safely)

## Goals / Non-Goals

**Goals:**
- Convert all level 5 village committee records to residential area entries
- Establish correct hierarchical relationships (city_id, county_id, street_id)
- Transform village names to standard format (remove "委会" suffix)
- Generate unique codes for each residential area
- Provide clear conversion statistics and error reporting

**Non-Goals:**
- Modifying existing residential area records
- Syncing future changes from administrative division to residential areas (one-time conversion)
- Validating administrative division data quality (assumes data is correct)
- Handling level 5 records that are not village committees (assumes all level 5 are villages)

## Decisions

### Decision 1: One-time batch script vs. ongoing sync service
**Choice:** One-time batch script

**Rationale:**
- Village administrative divisions change infrequently
- One-time conversion is simpler and has lower maintenance cost
- Future village updates can be handled through existing residential area management interface
- Avoids complexity of change detection and bidirectional sync

**Alternatives considered:**
- Ongoing sync service: Rejected due to complexity and low update frequency
- Database trigger: Rejected because administrative division is authoritative source, not residential area

### Decision 2: Code generation strategy
**Choice:** Append 4-digit sequence to village committee code

**Rationale:**
- Village committee code (12 digits) provides geographic context
- 4-digit sequence allows up to 9999 residential areas per village (sufficient for villages)
- Maintains sortability and readability
- Avoids collision with existing codes from other sources

**Alternatives considered:**
- UUID: Rejected due to loss of geographic meaning and sortability
- Auto-increment ID: Rejected because code field is string and should encode location

### Decision 3: Deduplication strategy
**Choice:** Check name + county_id combination before insertion

**Rationale:**
- Name + county_id uniquely identifies a residential area within administrative scope
- Prevents duplicate entries if script is run multiple times
- Simple to implement with single query per record

**Alternatives considered:**
- Check by code: Rejected because codes are generated and may differ across runs
- Check by name only: Rejected because same village name can exist in different counties
- No deduplication: Rejected because script should be idempotent

### Decision 4: Hierarchical ID resolution approach
**Choice:** Recursive parent_id traversal with level validation

**Rationale:**
- Administrative division table has complete parent_id chain
- Level field allows validation of each hierarchy level
- Handles missing intermediate levels gracefully (sets to NULL)
- Single query per level is acceptable for one-time batch operation

**Alternatives considered:**
- Pre-build lookup table: Rejected as premature optimization for one-time script
- Assume fixed hierarchy: Rejected because some regions may have incomplete data

### Decision 5: Error handling strategy
**Choice:** Log errors and continue processing remaining records

**Rationale:**
- Partial success is better than all-or-nothing for data migration
- Allows manual review and correction of problematic records
- Provides clear statistics on success/skip/error counts

**Alternatives considered:**
- Fail fast on first error: Rejected because one bad record shouldn't block thousands of good records
- Transaction rollback on any error: Rejected for same reason

## Risks / Trade-offs

**Risk:** Village committee names without "委会" suffix may not be transformed correctly
→ **Mitigation:** Name transformation logic checks for suffix presence before removal; logs all transformations for review

**Risk:** Broken parent_id chains in administrative division data will cause records to be skipped
→ **Mitigation:** Error logging includes specific village name and missing parent_id; manual data fix can be applied before re-running script

**Risk:** Generated codes may collide with manually entered residential area codes
→ **Mitigation:** Use 16-digit format (12-digit admin code + 4-digit sequence) which differs from typical manual entry patterns; database unique constraint will catch collisions

**Risk:** Large dataset may cause memory issues if loading all records at once
→ **Mitigation:** Process records in batches (e.g., 1000 records per batch) with periodic commit

**Trade-off:** Setting submission_status=2 (approved) bypasses review workflow
→ **Justification:** Administrative division data is authoritative and pre-validated; manual review of thousands of villages is impractical

**Trade-off:** One-time conversion means future administrative division changes won't auto-sync
→ **Justification:** Village changes are rare; manual updates through residential area interface are acceptable

## Migration Plan

**Execution Steps:**
1. Create standalone Go script in `services/masterdata/cmd/convert-villages/`
2. Test on development database with small dataset (single county)
3. Backup production `md_residential_area` table
4. Run conversion script on production database
5. Verify conversion statistics and sample records
6. Spot-check residential area query interface for village data

**Rollback Strategy:**
- If critical errors discovered: Delete records where `data_source = 2` and `community_type = 2`
- Script is idempotent: Can be re-run after fixing data issues

**Validation:**
- Compare total count of level 5 records vs. inserted residential areas
- Verify hierarchical IDs match administrative division structure
- Check name transformations for correctness
- Test residential area queries filtering by community_type=2

## Open Questions

None - design is complete and ready for implementation.
