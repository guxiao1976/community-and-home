## ADDED Requirements

### Requirement: System shall perform periodic health checks
The system SHALL perform health checks on all enabled models every 5 minutes.

#### Scenario: Scheduled health check
- **WHEN** health check scheduler runs
- **THEN** system sends test request to each enabled model and records result

#### Scenario: Skip disabled models
- **WHEN** health check runs
- **THEN** system skips models with enabled=false

#### Scenario: Skip recently checked models
- **WHEN** model was checked less than 4 minutes ago
- **THEN** system skips redundant check

### Requirement: System shall record health check results
The system SHALL record every health check result in am_health_check table.

#### Scenario: Record successful check
- **WHEN** health check succeeds
- **THEN** system records status="healthy", response_time_ms, and check_time

#### Scenario: Record failed check
- **WHEN** health check fails
- **THEN** system records status="unhealthy", error_message, and check_time

### Requirement: System shall calculate health status
The system SHALL calculate overall health status based on recent check history.

#### Scenario: Model is healthy
- **WHEN** last 10 checks have success rate >= 95%
- **THEN** system sets health_status="healthy"

#### Scenario: Model is degraded
- **WHEN** last 10 checks have success rate between 80% and 95%
- **THEN** system sets health_status="degraded"

#### Scenario: Model is unhealthy
- **WHEN** last 10 checks have success rate < 80%
- **THEN** system sets health_status="unhealthy"

#### Scenario: Insufficient check history
- **WHEN** model has fewer than 3 checks
- **THEN** system sets health_status="unknown"

### Requirement: System shall trigger alerts on health changes
The system SHALL trigger alerts when model health status changes.

#### Scenario: Health degrades to unhealthy
- **WHEN** model transitions from healthy to unhealthy
- **THEN** system creates alert record and logs warning

#### Scenario: Health recovers
- **WHEN** model transitions from unhealthy to healthy
- **THEN** system creates alert record with recovery message

#### Scenario: Consecutive failures
- **WHEN** model fails 3 consecutive health checks
- **THEN** system triggers immediate alert

### Requirement: Admin can manually trigger health check
The system SHALL allow administrators to manually trigger health check for a specific model.

#### Scenario: Manual check on healthy model
- **WHEN** admin triggers health check on model
- **THEN** system performs immediate check and returns result

#### Scenario: Manual check on disabled model
- **WHEN** admin triggers health check on disabled model
- **THEN** system performs check anyway (for testing purposes)

### Requirement: System shall expose health metrics
The system SHALL expose health metrics for monitoring systems.

#### Scenario: Get current health status
- **WHEN** monitoring system queries health endpoint
- **THEN** system returns current status of all models with last_check_time

#### Scenario: Get health history
- **WHEN** admin requests health history for a model
- **THEN** system returns last 100 check results with timestamps

### Requirement: System shall measure response time
The system SHALL measure and record response time for each health check.

#### Scenario: Fast response
- **WHEN** model responds in 500ms
- **THEN** system records response_time_ms=500

#### Scenario: Slow response
- **WHEN** model responds in 5000ms
- **THEN** system records response_time_ms=5000 and logs warning

#### Scenario: Timeout
- **WHEN** model does not respond within 10 seconds
- **THEN** system records status="unhealthy" with error="timeout"

### Requirement: System shall clean up old health records
The system SHALL automatically delete health check records older than 30 days.

#### Scenario: Daily cleanup
- **WHEN** cleanup job runs daily
- **THEN** system deletes records where check_time < NOW() - 30 days

#### Scenario: Preserve recent records
- **WHEN** cleanup job runs
- **THEN** system keeps all records from last 30 days
