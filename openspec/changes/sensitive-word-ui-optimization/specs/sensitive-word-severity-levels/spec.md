## ADDED Requirements

### Requirement: Two-tier severity level system
The system SHALL support exactly two severity levels for sensitive words: 违规 (violation) and 可疑 (suspicious).

#### Scenario: Display severity levels in query form
- **WHEN** user opens the sensitive word management page
- **THEN** the query form displays severity level options: 违规 and 可疑

#### Scenario: Filter by severity level
- **WHEN** user selects a severity level (违规 or 可疑) in the query form
- **THEN** the system filters sensitive words by the selected severity level

#### Scenario: Clear severity filter
- **WHEN** user clears the severity level selection
- **THEN** the system shows all sensitive words regardless of severity level

### Requirement: Severity level in add form
The add sensitive word form SHALL provide severity level selection with only two options: 违规 (violation) and 可疑 (suspicious).

#### Scenario: Display severity options in add form
- **WHEN** user opens the add sensitive word form
- **THEN** the form displays severity level options: 违规 and 可疑

#### Scenario: Select severity when adding
- **WHEN** user selects a severity level (违规 or 可疑) and submits the form
- **THEN** the system creates a new sensitive word with the selected severity level

#### Scenario: Severity is required
- **WHEN** user attempts to submit the add form without selecting a severity level
- **THEN** the system displays a validation error requiring severity level selection

### Requirement: Remove handling action field
The add sensitive word form SHALL NOT include a "处理动作" (handling action) field.

#### Scenario: Add form without handling action
- **WHEN** user opens the add sensitive word form
- **THEN** the form does not display a "处理动作" field

#### Scenario: Submit without handling action
- **WHEN** user submits the add form with category and severity level
- **THEN** the system successfully creates the sensitive word without requiring a handling action

### Requirement: Severity level validation
The system SHALL only accept 违规 or 可疑 as valid severity level values.

#### Scenario: Validate severity on submission
- **WHEN** user submits a sensitive word with a severity level
- **THEN** the system validates that the severity is either 违规 or 可疑

#### Scenario: Reject invalid severity
- **WHEN** an invalid severity value is submitted
- **THEN** the system rejects the submission with a validation error
