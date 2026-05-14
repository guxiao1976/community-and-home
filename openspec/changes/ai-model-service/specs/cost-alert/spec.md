## ADDED Requirements

### Requirement: Admin can configure cost alert thresholds
The system SHALL allow administrators to configure daily and monthly cost alert thresholds.

#### Scenario: Set daily threshold
- **WHEN** admin sets daily_threshold=100 USD
- **THEN** system triggers alert when daily cost exceeds 100 USD

#### Scenario: Set monthly threshold
- **WHEN** admin sets monthly_threshold=3000 USD
- **THEN** system triggers alert when monthly cost exceeds 3000 USD

#### Scenario: Set warning threshold
- **WHEN** admin sets warning_threshold=80 percent
- **THEN** system triggers warning when cost reaches 80% of threshold

### Requirement: System shall monitor cost in real-time
The system SHALL track cumulative cost and check against thresholds after each call.

#### Scenario: Cost within threshold
- **WHEN** daily cost is 50 USD and threshold is 100 USD
- **THEN** system allows calls to proceed normally

#### Scenario: Cost exceeds threshold
- **WHEN** daily cost reaches 101 USD and threshold is 100 USD
- **THEN** system triggers alert and optionally blocks further calls

#### Scenario: Cost reaches warning level
- **WHEN** daily cost reaches 80 USD and threshold is 100 USD
- **THEN** system triggers warning alert

### Requirement: System shall create alert records
The system SHALL create records in am_alert_record table when thresholds are exceeded.

#### Scenario: Create daily alert
- **WHEN** daily cost exceeds threshold
- **THEN** system creates alert with alert_type="daily_cost_exceeded", threshold_value, actual_value, alert_time

#### Scenario: Create monthly alert
- **WHEN** monthly cost exceeds threshold
- **THEN** system creates alert with alert_type="monthly_cost_exceeded"

#### Scenario: Create warning alert
- **WHEN** cost reaches warning threshold
- **THEN** system creates alert with alert_type="cost_warning", severity="warning"

### Requirement: System shall prevent duplicate alerts
The system SHALL NOT create duplicate alerts for the same threshold breach.

#### Scenario: First breach triggers alert
- **WHEN** daily cost first exceeds threshold
- **THEN** system creates alert record

#### Scenario: Subsequent calls do not trigger duplicate
- **WHEN** daily cost continues to exceed threshold
- **THEN** system does not create additional alerts for same day

#### Scenario: Reset on new day
- **WHEN** new day starts
- **THEN** system resets alert state and can trigger new alerts

### Requirement: System shall support alert notifications
The system SHALL support multiple notification channels for alerts.

#### Scenario: Log notification
- **WHEN** alert is triggered
- **THEN** system logs alert to application log

#### Scenario: Database notification
- **WHEN** alert is triggered
- **THEN** system creates record in am_alert_record table

#### Scenario: Webhook notification (Phase 2)
- **WHEN** alert is triggered and webhook is configured
- **THEN** system sends POST request to configured webhook URL

### Requirement: System shall calculate cost by model
The system SHALL track and alert on per-model cost thresholds.

#### Scenario: Model-specific threshold
- **WHEN** admin sets threshold for specific model_id
- **THEN** system monitors cost for that model only

#### Scenario: Model exceeds threshold
- **WHEN** specific model's daily cost exceeds its threshold
- **THEN** system triggers alert and optionally disables that model

### Requirement: System shall provide cost projection
The system SHALL project end-of-month cost based on current usage trend.

#### Scenario: Project monthly cost
- **WHEN** 10 days into month with 1000 USD spent
- **THEN** system projects monthly cost as (1000 / 10) * 30 = 3000 USD

#### Scenario: Alert on projected overage
- **WHEN** projected monthly cost exceeds threshold
- **THEN** system triggers early warning alert

### Requirement: Admin can view alert history
The system SHALL provide interface to view all triggered alerts.

#### Scenario: List recent alerts
- **WHEN** admin requests alert history
- **THEN** system returns alerts ordered by alert_time descending

#### Scenario: Filter by alert type
- **WHEN** admin filters by alert_type="daily_cost_exceeded"
- **THEN** system returns only daily cost alerts

#### Scenario: Filter by date range
- **WHEN** admin filters alerts for last 30 days
- **THEN** system returns alerts where alert_time >= NOW() - 30 days

### Requirement: Admin can acknowledge alerts
The system SHALL allow administrators to acknowledge and resolve alerts.

#### Scenario: Acknowledge alert
- **WHEN** admin acknowledges an alert
- **THEN** system sets acknowledged=true, acknowledged_by, acknowledged_at

#### Scenario: Add resolution notes
- **WHEN** admin acknowledges alert with notes
- **THEN** system stores notes in resolution_notes field

### Requirement: System shall support alert suppression
The system SHALL allow temporary suppression of alerts during maintenance.

#### Scenario: Suppress alerts for model
- **WHEN** admin enables alert suppression for model
- **THEN** system does not trigger alerts for that model until suppression is disabled

#### Scenario: Suppression expires
- **WHEN** suppression period ends
- **THEN** system automatically re-enables alerts
