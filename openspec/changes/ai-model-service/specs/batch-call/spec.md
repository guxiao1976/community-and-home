## ADDED Requirements

### Requirement: System shall support batch call merging
The system SHALL merge multiple items into a single prompt for batch processing.

#### Scenario: Merge 100 items into one prompt
- **WHEN** client calls CallModelBatch with 100 items
- **THEN** system constructs single prompt containing all items and makes one API call

#### Scenario: Use template for batch prompt
- **WHEN** batch call specifies template_id
- **THEN** system renders template with items array as variable

#### Scenario: Custom batch prompt
- **WHEN** batch call provides custom prompt with {{items}} placeholder
- **THEN** system replaces {{items}} with formatted item list

### Requirement: System shall enforce batch size limits
The system SHALL enforce maximum batch size to prevent oversized requests.

#### Scenario: Batch within limit
- **WHEN** client sends batch of 100 items (within limit)
- **THEN** system processes batch normally

#### Scenario: Batch exceeds limit
- **WHEN** client sends batch of 1000 items (exceeds limit of 500)
- **THEN** system returns error "Batch size exceeds maximum of 500"

#### Scenario: Auto-split large batches (Phase 2)
- **WHEN** client sends batch of 1000 items with auto_split=true
- **THEN** system splits into multiple batches of 500 and processes sequentially

### Requirement: System shall parse batch responses
The system SHALL parse model response and extract results for each item.

#### Scenario: Parse JSON array response
- **WHEN** model returns JSON array with results for each item
- **THEN** system parses array and maps results to input items

#### Scenario: Validate result count matches input
- **WHEN** model returns 95 results for 100 input items
- **THEN** system identifies 5 missing results and marks them as failed

#### Scenario: Handle malformed JSON
- **WHEN** model returns invalid JSON
- **THEN** system returns error "Failed to parse batch response"

### Requirement: System shall handle partial batch failures
The system SHALL support partial success when some items fail processing.

#### Scenario: Partial success
- **WHEN** model successfully processes 90 out of 100 items
- **THEN** system returns successful results and lists 10 failed items

#### Scenario: Complete failure
- **WHEN** model fails to process any items
- **THEN** system returns error with failure reason

#### Scenario: Retry failed items
- **WHEN** batch has partial failures and retry=true
- **THEN** system retries only failed items in new batch

### Requirement: System shall validate batch results
The system SHALL validate that batch results match input items.

#### Scenario: Validate item IDs match
- **WHEN** model returns results with item IDs
- **THEN** system verifies each result ID matches an input item ID

#### Scenario: Detect duplicate results
- **WHEN** model returns duplicate results for same item
- **THEN** system keeps first result and logs warning

#### Scenario: Detect missing results
- **WHEN** model omits results for some items
- **THEN** system marks missing items as failed with reason "no_result"

### Requirement: System shall estimate batch cost
The system SHALL estimate total cost before processing batch.

#### Scenario: Estimate before processing
- **WHEN** client requests batch processing
- **THEN** system estimates cost based on batch size and returns in metadata

#### Scenario: Compare estimated vs actual cost
- **WHEN** batch completes
- **THEN** system logs difference between estimated and actual cost

### Requirement: System shall support batch progress tracking
The system SHALL provide progress updates for long-running batch operations.

#### Scenario: Track batch progress
- **WHEN** batch is split into multiple API calls
- **THEN** system updates progress after each sub-batch completes

#### Scenario: Query batch status
- **WHEN** client queries batch status by request_id
- **THEN** system returns processed_count, total_count, and status

### Requirement: System shall support batch cancellation
The system SHALL allow clients to cancel in-progress batch operations.

#### Scenario: Cancel batch
- **WHEN** client cancels batch request
- **THEN** system stops processing remaining items and returns partial results

#### Scenario: Cancel between sub-batches
- **WHEN** batch is cancelled after first sub-batch completes
- **THEN** system returns results from completed sub-batch

### Requirement: System shall optimize batch formatting
The system SHALL format batch items efficiently to minimize token usage.

#### Scenario: Compact JSON formatting
- **WHEN** constructing batch prompt
- **THEN** system uses compact JSON without extra whitespace

#### Scenario: Numbered list formatting
- **WHEN** template uses numbered list format
- **THEN** system formats as "1. item1\n2. item2\n..."

#### Scenario: CSV formatting
- **WHEN** template uses CSV format
- **THEN** system formats as comma-separated values

### Requirement: System shall support batch result transformation
The system SHALL allow custom transformation of batch results.

#### Scenario: Extract specific fields
- **WHEN** batch response contains extra fields
- **THEN** system extracts only requested fields per configuration

#### Scenario: Apply default values
- **WHEN** result is missing optional field
- **THEN** system applies configured default value

### Requirement: System shall log batch operations
The system SHALL log batch operations with aggregated metrics.

#### Scenario: Log batch call
- **WHEN** batch completes
- **THEN** system logs single entry with batch_size, successful_count, failed_count, total_cost

#### Scenario: Link to individual items
- **WHEN** batch is logged
- **THEN** system stores business_ids array for traceability
