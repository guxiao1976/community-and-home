## 1. Project Setup

- [ ] 1.1 Create directory structure `services/masterdata/cmd/convert-villages/`
- [ ] 1.2 Create main.go with basic CLI structure and configuration loading
- [ ] 1.3 Add database connection setup using existing masterdata config

## 2. Data Extraction

- [ ] 2.1 Implement function to query all level=5 records from md_administrative_division
- [ ] 2.2 Add batch processing logic to handle records in chunks of 1000
- [ ] 2.3 Implement error handling for database query failures

## 3. Hierarchical ID Resolution

- [ ] 3.1 Implement function to resolve parent_id chain for a given village record
- [ ] 3.2 Add level validation to extract street_id (level 4), county_id (level 3), city_id (level 2)
- [ ] 3.3 Handle missing intermediate levels by setting NULL values
- [ ] 3.4 Add error logging for broken parent_id chains

## 4. Name Transformation

- [ ] 4.1 Implement function to remove "委会" suffix from village committee names
- [ ] 4.2 Add logic to preserve original name if suffix not found
- [ ] 4.3 Add unit tests for name transformation edge cases

## 5. Code Generation

- [ ] 5.1 Implement function to query existing residential area codes with given prefix
- [ ] 5.2 Add logic to find next available 4-digit sequence number
- [ ] 5.3 Generate 16-digit code by appending sequence to village committee code
- [ ] 5.4 Handle code collision with retry logic

## 6. Deduplication Check

- [ ] 6.1 Implement function to check if residential area exists by name and county_id
- [ ] 6.2 Add logic to skip insertion if duplicate found
- [ ] 6.3 Increment skipped count in statistics

## 7. Data Insertion

- [ ] 7.1 Implement function to insert residential area record with all required fields
- [ ] 7.2 Set community_type=2, submission_status=2, data_source=2
- [ ] 7.3 Set created_at and updated_at to current timestamp
- [ ] 7.4 Handle database constraint violations with error logging

## 8. Statistics and Logging

- [ ] 8.1 Implement statistics tracking (total, success, skipped, error counts)
- [ ] 8.2 Add detailed logging for each processing step
- [ ] 8.3 Log sample error messages for failed records
- [ ] 8.4 Output final conversion summary to console

## 9. Testing

- [ ] 9.1 Test on development database with single county dataset
- [ ] 9.2 Verify hierarchical ID resolution correctness
- [ ] 9.3 Verify name transformation results
- [ ] 9.4 Verify code generation uniqueness
- [ ] 9.5 Test idempotency by running script twice

## 10. Documentation

- [ ] 10.1 Add README.md with script usage instructions
- [ ] 10.2 Document command-line flags and configuration options
- [ ] 10.3 Add example output and troubleshooting guide
