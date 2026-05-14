## ADDED Requirements

### Requirement: System shall aggregate daily statistics
The system SHALL aggregate call logs into daily statistics every midnight.

#### Scenario: Daily aggregation runs
- **WHEN** aggregation job runs at 00:00
- **THEN** system calculates statistics for previous day and stores in am_usage_statistics

#### Scenario: Aggregate by model
- **WHEN** aggregation runs
- **THEN** system creates separate record for each model_id

#### Scenario: Aggregate by business type
- **WHEN** aggregation runs
- **THEN** system creates separate record for each business_type

### Requirement: System shall calculate call metrics
The system SHALL calculate total calls, successful calls, failed calls, and success rate.

#### Scenario: Calculate call counts
- **WHEN** aggregating daily statistics
- **THEN** system counts total_calls, successful_calls (status='success'), failed_calls (status='failed')

#### Scenario: Calculate success rate
- **WHEN** aggregating statistics
- **THEN** system calculates success_rate = successful_calls / total_calls * 100

### Requirement: System shall calculate token usage
The system SHALL aggregate input tokens, output tokens, and total tokens.

#### Scenario: Sum token usage
- **WHEN** aggregating daily statistics
- **THEN** system sums total_input_tokens, total_output_tokens, total_tokens

#### Scenario: Calculate average tokens per call
- **WHEN** aggregating statistics
- **THEN** system calculates avg_tokens_per_call = total_tokens / total_calls

### Requirement: System shall calculate cost metrics
The system SHALL aggregate total cost and calculate cost per call.

#### Scenario: Sum daily cost
- **WHEN** aggregating daily statistics
- **THEN** system sums total_cost from all calls

#### Scenario: Calculate average cost per call
- **WHEN** aggregating statistics
- **THEN** system calculates avg_cost_per_call = total_cost / total_calls

#### Scenario: Calculate cost by business type
- **WHEN** aggregating statistics
- **THEN** system breaks down cost by business_type

### Requirement: System shall calculate performance metrics
The system SHALL aggregate response time statistics.

#### Scenario: Calculate average response time
- **WHEN** aggregating statistics
- **THEN** system calculates avg_duration_ms from all calls

#### Scenario: Calculate P95 response time
- **WHEN** aggregating statistics
- **THEN** system calculates p95_duration_ms (95th percentile)

#### Scenario: Calculate P99 response time
- **WHEN** aggregating statistics
- **THEN** system calculates p99_duration_ms (99th percentile)

### Requirement: Admin can query statistics by date range
The system SHALL provide API to query statistics for specified date range.

#### Scenario: Query last 7 days
- **WHEN** admin queries statistics for last 7 days
- **THEN** system returns aggregated data for each day

#### Scenario: Query specific month
- **WHEN** admin queries statistics for 2026-05
- **THEN** system returns daily statistics for May 2026

#### Scenario: Query year-to-date
- **WHEN** admin queries YTD statistics
- **THEN** system returns aggregated data from Jan 1 to current date

### Requirement: Admin can query statistics by model
The system SHALL provide API to query statistics filtered by model.

#### Scenario: Query single model statistics
- **WHEN** admin queries statistics for model_id=1
- **THEN** system returns statistics for that model only

#### Scenario: Compare models
- **WHEN** admin queries statistics for multiple model_ids
- **THEN** system returns comparative statistics for each model

### Requirement: Admin can query statistics by business type
The system SHALL provide API to query statistics filtered by business type.

#### Scenario: Query by business type
- **WHEN** admin queries statistics for business_type="sensitive_word_classify"
- **THEN** system returns statistics for that business type only

#### Scenario: Business type breakdown
- **WHEN** admin queries overall statistics
- **THEN** system includes breakdown by business_type

### Requirement: System shall provide trend analysis
The system SHALL calculate trends comparing current period to previous period.

#### Scenario: Week-over-week trend
- **WHEN** admin queries weekly trend
- **THEN** system compares current week to previous week and shows percentage change

#### Scenario: Month-over-month trend
- **WHEN** admin queries monthly trend
- **THEN** system compares current month to previous month

### Requirement: System shall visualize statistics
The system SHALL provide data formatted for chart visualization.

#### Scenario: Time series data
- **WHEN** admin requests chart data
- **THEN** system returns array of {date, total_calls, total_cost} for plotting

#### Scenario: Model comparison chart
- **WHEN** admin requests model comparison
- **THEN** system returns data grouped by model for bar chart

### Requirement: System shall export statistics
The system SHALL allow administrators to export statistics in CSV format.

#### Scenario: Export daily statistics
- **WHEN** admin exports statistics for date range
- **THEN** system generates CSV with all metrics

#### Scenario: Export includes all fields
- **WHEN** exporting statistics
- **THEN** CSV includes date, model, business_type, calls, tokens, cost, success_rate, avg_duration

### Requirement: System shall calculate cost projections
The system SHALL project future costs based on historical trends.

#### Scenario: Project monthly cost
- **WHEN** admin requests monthly projection
- **THEN** system calculates based on daily average and remaining days

#### Scenario: Project with growth rate
- **WHEN** usage is growing 10% week-over-week
- **THEN** system factors growth rate into projection
