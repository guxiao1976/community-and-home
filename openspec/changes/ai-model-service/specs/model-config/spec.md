## ADDED Requirements

### Requirement: Admin can create model configuration
The system SHALL allow administrators to create a new AI model configuration with provider type, model name, API endpoint, and API key.

#### Scenario: Create Claude model configuration
- **WHEN** admin submits a model configuration with provider="claude", model_name="claude-opus-4", api_endpoint="https://api.anthropic.com", and api_key="sk-ant-xxx"
- **THEN** system creates the configuration and returns the model ID

#### Scenario: Create configuration with invalid provider
- **WHEN** admin submits a model configuration with unsupported provider type
- **THEN** system returns error "Unsupported provider type"

#### Scenario: Create configuration with duplicate name
- **WHEN** admin submits a model configuration with a name that already exists
- **THEN** system returns error "Model name already exists"

### Requirement: Admin can update model configuration
The system SHALL allow administrators to update existing model configuration including endpoint, API key, priority, and enabled status.

#### Scenario: Update model API key
- **WHEN** admin updates the API key of an existing model
- **THEN** system updates the configuration and invalidates cached credentials

#### Scenario: Update disabled model
- **WHEN** admin attempts to update a disabled model
- **THEN** system allows the update but keeps the model disabled

### Requirement: Admin can delete model configuration
The system SHALL allow administrators to soft-delete a model configuration.

#### Scenario: Delete unused model
- **WHEN** admin deletes a model that has no active calls
- **THEN** system marks the model as deleted with delete_time timestamp

#### Scenario: Delete model with active calls
- **WHEN** admin deletes a model that has active calls in the last 24 hours
- **THEN** system shows warning but allows deletion

### Requirement: Admin can enable or disable model
The system SHALL allow administrators to enable or disable a model without deleting it.

#### Scenario: Disable unhealthy model
- **WHEN** admin disables a model
- **THEN** system stops routing new requests to this model

#### Scenario: Enable previously disabled model
- **WHEN** admin enables a previously disabled model
- **THEN** system includes the model in routing decisions

### Requirement: Admin can list all model configurations
The system SHALL return a list of all model configurations with their status and health information.

#### Scenario: List all models
- **WHEN** admin requests model list
- **THEN** system returns all models with id, name, provider, status, health_status, and last_check_time

#### Scenario: Filter models by provider
- **WHEN** admin requests model list with provider filter
- **THEN** system returns only models matching the specified provider

### Requirement: Admin can test model connectivity
The system SHALL allow administrators to test a model's connectivity before saving or after configuration.

#### Scenario: Test valid model configuration
- **WHEN** admin tests a model with valid credentials
- **THEN** system sends a test request and returns success with response time

#### Scenario: Test invalid API key
- **WHEN** admin tests a model with invalid API key
- **THEN** system returns error "Authentication failed"

### Requirement: System shall encrypt API keys
The system SHALL encrypt all API keys using AES-256 before storing in database.

#### Scenario: Store new API key
- **WHEN** system stores a new model configuration
- **THEN** API key is encrypted and only encrypted value is stored

#### Scenario: Retrieve API key for use
- **WHEN** system needs to call a model
- **THEN** API key is decrypted in memory and never logged

### Requirement: System shall support model priority
The system SHALL support priority levels for models to control routing preferences.

#### Scenario: Route to highest priority model
- **WHEN** multiple models are available for the same provider
- **THEN** system routes requests to the model with highest priority (lowest number)

#### Scenario: Fallback to lower priority model
- **WHEN** highest priority model is unhealthy
- **THEN** system routes to next available model by priority

### Requirement: System shall support model quotas
The system SHALL allow administrators to set daily and monthly quotas for each model.

#### Scenario: Enforce daily quota
- **WHEN** a model reaches its daily quota limit
- **THEN** system stops routing requests to this model until next day

#### Scenario: Warn approaching quota
- **WHEN** a model reaches 80% of its quota
- **THEN** system logs a warning and triggers alert
