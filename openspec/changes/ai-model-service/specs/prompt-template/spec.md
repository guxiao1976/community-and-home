## ADDED Requirements

### Requirement: Admin can create prompt template
The system SHALL allow administrators to create reusable prompt templates with variables.

#### Scenario: Create template with variables
- **WHEN** admin creates template with content "Classify these words: {{words}}"
- **THEN** system stores template with name, category, and content

#### Scenario: Create template without variables
- **WHEN** admin creates template with static content
- **THEN** system stores template and marks has_variables=false

#### Scenario: Duplicate template name
- **WHEN** admin creates template with existing name
- **THEN** system returns error "Template name already exists"

### Requirement: System shall validate template syntax
The system SHALL validate template syntax and variable placeholders.

#### Scenario: Valid template syntax
- **WHEN** admin creates template with {{variable}} syntax
- **THEN** system validates and extracts variable names

#### Scenario: Invalid variable syntax
- **WHEN** admin creates template with {variable} (single brace)
- **THEN** system returns error "Invalid variable syntax, use {{variable}}"

#### Scenario: Unclosed variable
- **WHEN** admin creates template with {{variable
- **THEN** system returns error "Unclosed variable placeholder"

### Requirement: System shall render template with variables
The system SHALL replace variable placeholders with actual values when rendering.

#### Scenario: Render with all variables provided
- **WHEN** template "Hello {{name}}" is rendered with name="World"
- **THEN** system returns "Hello World"

#### Scenario: Render with missing variable
- **WHEN** template "Hello {{name}}" is rendered without name
- **THEN** system returns error "Missing required variable: name"

#### Scenario: Render with extra variables
- **WHEN** template is rendered with extra unused variables
- **THEN** system ignores extra variables and renders successfully

### Requirement: System shall support template versioning
The system SHALL maintain version history for templates.

#### Scenario: Create initial version
- **WHEN** admin creates new template
- **THEN** system sets version=1

#### Scenario: Update creates new version
- **WHEN** admin updates template content
- **THEN** system increments version and keeps old version in history

#### Scenario: Rollback to previous version
- **WHEN** admin rolls back to version 2
- **THEN** system creates new version with content from version 2

### Requirement: System shall categorize templates
The system SHALL organize templates by category for easy discovery.

#### Scenario: Create template with category
- **WHEN** admin creates template with category="sensitive_word_classify"
- **THEN** system stores category and allows filtering by it

#### Scenario: List templates by category
- **WHEN** admin requests templates for category="content_moderation"
- **THEN** system returns only templates in that category

### Requirement: System shall track template usage
The system SHALL track how many times each template is used.

#### Scenario: Increment usage count
- **WHEN** template is rendered for a call
- **THEN** system increments usage_count

#### Scenario: Track last used time
- **WHEN** template is rendered
- **THEN** system updates last_used_at timestamp

### Requirement: Admin can test template rendering
The system SHALL provide interface to test template rendering with sample data.

#### Scenario: Test with sample variables
- **WHEN** admin tests template with sample variable values
- **THEN** system renders template and returns preview

#### Scenario: Test reveals missing variable
- **WHEN** admin tests template without required variable
- **THEN** system shows error highlighting missing variable

### Requirement: System shall support template variables with defaults
The system SHALL allow optional variables with default values.

#### Scenario: Variable with default
- **WHEN** template contains "{{name:Guest}}" and name is not provided
- **THEN** system uses "Guest" as default value

#### Scenario: Variable with default overridden
- **WHEN** template contains "{{name:Guest}}" and name="Alice" is provided
- **THEN** system uses "Alice" instead of default

### Requirement: System shall validate rendered output
The system SHALL validate that rendered template meets length requirements.

#### Scenario: Rendered template within limits
- **WHEN** rendered template is 5000 characters
- **THEN** system accepts and uses template

#### Scenario: Rendered template exceeds limit
- **WHEN** rendered template exceeds 100000 characters
- **THEN** system returns error "Rendered template exceeds maximum length"

### Requirement: Admin can enable or disable templates
The system SHALL allow administrators to enable or disable templates.

#### Scenario: Disable unused template
- **WHEN** admin disables a template
- **THEN** system sets enabled=false and template cannot be used

#### Scenario: Enable template
- **WHEN** admin enables a disabled template
- **THEN** system sets enabled=true and template becomes available

### Requirement: System shall support template cloning
The system SHALL allow administrators to clone existing templates.

#### Scenario: Clone template
- **WHEN** admin clones template "A" as "B"
- **THEN** system creates new template "B" with same content as "A" at version 1

#### Scenario: Clone and modify
- **WHEN** admin clones and immediately modifies
- **THEN** system creates new template with modified content
