## ADDED Requirements

### Requirement: Service shall provide CallModel RPC method
The system SHALL provide a CallModel RPC method for single AI model invocations.

#### Scenario: Successful single call
- **WHEN** client calls CallModel with valid prompt and model_id
- **THEN** system routes to specified model, returns response text and token usage

#### Scenario: Call with automatic model selection
- **WHEN** client calls CallModel with provider type but no model_id
- **THEN** system selects best available model based on health and priority

#### Scenario: Call to disabled model
- **WHEN** client calls CallModel with a disabled model_id
- **THEN** system returns error "Model is disabled"

#### Scenario: Call with timeout
- **WHEN** model response exceeds configured timeout (60s default)
- **THEN** system cancels request and returns error "Request timeout"

### Requirement: Service shall provide CallModelBatch RPC method
The system SHALL provide a CallModelBatch RPC method for batch AI model invocations.

#### Scenario: Successful batch call
- **WHEN** client calls CallModelBatch with 100 items and a prompt template
- **THEN** system merges items into single prompt, calls model once, returns parsed results

#### Scenario: Batch call with partial failure
- **WHEN** model returns results for only 80 out of 100 items
- **THEN** system returns successful results and lists failed items

#### Scenario: Batch size exceeds limit
- **WHEN** client calls CallModelBatch with more than configured max_batch_size
- **THEN** system returns error "Batch size exceeds limit"

#### Scenario: Batch call with invalid JSON response
- **WHEN** model returns non-JSON or malformed JSON
- **THEN** system returns error "Failed to parse model response"

### Requirement: Service shall provide GetAvailableModels RPC method
The system SHALL provide a GetAvailableModels RPC method to query available models.

#### Scenario: Get all available models
- **WHEN** client calls GetAvailableModels with no filter
- **THEN** system returns all enabled and healthy models

#### Scenario: Filter by provider
- **WHEN** client calls GetAvailableModels with provider="claude"
- **THEN** system returns only Claude models that are enabled and healthy

#### Scenario: No available models
- **WHEN** all models are disabled or unhealthy
- **THEN** system returns empty list with warning message

### Requirement: Service shall provide HealthCheck RPC method
The system SHALL provide a HealthCheck RPC method for service health verification.

#### Scenario: Service is healthy
- **WHEN** client calls HealthCheck
- **THEN** system returns status="healthy" with service version and uptime

#### Scenario: Database connection failed
- **WHEN** client calls HealthCheck and database is unreachable
- **THEN** system returns status="unhealthy" with error details

### Requirement: Service shall log all RPC calls
The system SHALL log every RPC call with request details, response, duration, and cost.

#### Scenario: Log successful call
- **WHEN** CallModel completes successfully
- **THEN** system writes log entry with request_id, model_id, prompt_length, response_length, tokens_used, cost, and duration

#### Scenario: Log failed call
- **WHEN** CallModel fails with error
- **THEN** system writes log entry with error_message and error_code

### Requirement: Service shall implement retry mechanism
The system SHALL retry failed model calls up to 3 times with exponential backoff.

#### Scenario: Retry on transient error
- **WHEN** model returns 429 rate limit error
- **THEN** system waits and retries up to 3 times

#### Scenario: No retry on authentication error
- **WHEN** model returns 401 authentication error
- **THEN** system does not retry and returns error immediately

#### Scenario: Retry exhausted
- **WHEN** all 3 retry attempts fail
- **THEN** system returns error "Max retries exceeded"

### Requirement: Service shall validate request parameters
The system SHALL validate all RPC request parameters before processing.

#### Scenario: Missing required field
- **WHEN** client calls CallModel without prompt
- **THEN** system returns error "Missing required field: prompt"

#### Scenario: Prompt exceeds max length
- **WHEN** client calls CallModel with prompt longer than 100000 characters
- **THEN** system returns error "Prompt exceeds maximum length"

#### Scenario: Invalid temperature value
- **WHEN** client calls CallModel with temperature outside 0-1 range
- **THEN** system returns error "Temperature must be between 0 and 1"

### Requirement: Service shall support request cancellation
The system SHALL support context cancellation for long-running requests.

#### Scenario: Client cancels request
- **WHEN** client cancels context during model call
- **THEN** system stops processing and returns context.Canceled error

#### Scenario: Server shutdown during request
- **WHEN** server receives shutdown signal during active request
- **THEN** system completes in-flight requests within grace period (30s)

### Requirement: Service shall estimate cost before calling
The system SHALL estimate the cost of a request before making the actual API call.

#### Scenario: Estimate single call cost
- **WHEN** client calls CallModel
- **THEN** system estimates cost based on prompt tokens and model pricing

#### Scenario: Estimate batch call cost
- **WHEN** client calls CallModelBatch with 100 items
- **THEN** system estimates total cost and includes in response metadata

### Requirement: Service shall enforce rate limiting
The system SHALL enforce per-client rate limiting to prevent abuse.

#### Scenario: Within rate limit
- **WHEN** client makes 10 requests per minute (within limit)
- **THEN** all requests are processed normally

#### Scenario: Exceeds rate limit
- **WHEN** client makes 100 requests per minute (exceeds limit)
- **THEN** system returns error "Rate limit exceeded" for excess requests
