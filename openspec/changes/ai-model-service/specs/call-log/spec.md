## ADDED Requirements

### Requirement: System shall log every model call
The system SHALL create a log entry in am_call_log table for every model API call.

#### Scenario: Log successful call
- **WHEN** model call completes successfully
- **THEN** system records request_id, model_id, business_type, prompt_length, response_length, input_tokens, output_tokens, cost, duration_ms, status="success"

#### Scenario: Log failed call
- **WHEN** model call fails
- **THEN** system records request_id, model_id, error_message, error_code, status="failed"

#### Scenario: Log cancelled call
- **WHEN** client cancels request
- **THEN** system records status="cancelled" with partial duration

### Requirement: System shall generate unique request ID
The system SHALL generate a unique request_id for each call for tracing.

#### Scenario: Generate request ID
- **WHEN** new call is initiated
- **THEN** system generates UUID v4 as request_id

#### Scenario: Include request ID in response
- **WHEN** call completes
- **THEN** system includes request_id in response metadata

### Requirement: System shall record business context
The system SHALL record business_type and business_id to track call purpose.

#### Scenario: Log with business context
- **WHEN** Masterdata calls for sensitive word classification
- **THEN** system records business_type="sensitive_word_classify", business_id=<word_id>

#### Scenario: Log without business context
- **WHEN** call does not specify business context
- **THEN** system records business_type=null, business_id=null

### Requirement: System shall calculate and record cost
The system SHALL calculate actual cost based on token usage and record in log.

#### Scenario: Record cost for Claude call
- **WHEN** Claude call uses 1000 input tokens and 500 output tokens
- **THEN** system calculates cost and records in cost field

#### Scenario: Record zero cost for Ollama
- **WHEN** Ollama call completes
- **THEN** system records cost=0

### Requirement: System shall record timing information
The system SHALL record call start time, end time, and duration.

#### Scenario: Record timing
- **WHEN** call starts at T0 and completes at T1
- **THEN** system records created_at=T0, duration_ms=(T1-T0)

#### Scenario: Record timeout
- **WHEN** call times out after 60 seconds
- **THEN** system records duration_ms=60000, status="failed", error_message="timeout"

### Requirement: System shall support log querying
The system SHALL provide API to query call logs with filters.

#### Scenario: Query by date range
- **WHEN** admin queries logs for last 7 days
- **THEN** system returns logs where created_at >= NOW() - 7 days

#### Scenario: Query by model
- **WHEN** admin queries logs for specific model_id
- **THEN** system returns logs for that model only

#### Scenario: Query by business type
- **WHEN** admin queries logs for business_type="sensitive_word_classify"
- **THEN** system returns logs for that business type only

#### Scenario: Query by status
- **WHEN** admin queries failed calls
- **THEN** system returns logs where status="failed"

### Requirement: System shall aggregate log statistics
The system SHALL aggregate logs into daily statistics in am_usage_statistics table.

#### Scenario: Daily aggregation
- **WHEN** aggregation job runs at midnight
- **THEN** system calculates total_calls, successful_calls, failed_calls, total_tokens, total_cost for previous day

#### Scenario: Aggregate by model
- **WHEN** aggregation runs
- **THEN** system creates separate statistics record for each model

#### Scenario: Aggregate by business type
- **WHEN** aggregation runs
- **THEN** system includes breakdown by business_type in statistics

### Requirement: System shall clean up old logs
The system SHALL automatically archive or delete logs older than 90 days.

#### Scenario: Archive old logs
- **WHEN** cleanup job runs monthly
- **THEN** system moves logs older than 90 days to archive table

#### Scenario: Delete archived logs
- **WHEN** cleanup job runs
- **THEN** system deletes archived logs older than 1 year

### Requirement: System shall protect sensitive data in logs
The system SHALL NOT log full prompts or responses containing sensitive data.

#### Scenario: Log prompt length only
- **WHEN** logging a call
- **THEN** system records prompt_length but not full prompt text

#### Scenario: Log response length only
- **WHEN** logging a call
- **THEN** system records response_length but not full response text

#### Scenario: Sanitize error messages
- **WHEN** error message contains API key
- **THEN** system redacts API key before logging

### Requirement: System shall support log export
The system SHALL allow administrators to export logs in CSV format.

#### Scenario: Export filtered logs
- **WHEN** admin exports logs with date range filter
- **THEN** system generates CSV file with selected logs

#### Scenario: Export includes all fields
- **WHEN** admin exports logs
- **THEN** CSV includes request_id, model_id, business_type, tokens, cost, duration, status, created_at
