## ADDED Requirements

### Requirement: System shall provide unified ModelAdapter interface
The system SHALL define a unified ModelAdapter interface that abstracts different AI provider APIs.

#### Scenario: Call method signature
- **WHEN** any adapter implements the Call method
- **THEN** it accepts CallRequest and returns CallResponse with consistent structure

#### Scenario: HealthCheck method signature
- **WHEN** any adapter implements HealthCheck method
- **THEN** it returns error if model is unreachable or nil if healthy

#### Scenario: EstimateCost method signature
- **WHEN** any adapter implements EstimateCost method
- **THEN** it returns estimated cost in USD based on request parameters

### Requirement: System shall implement Claude adapter
The system SHALL implement a ClaudeAdapter that calls Anthropic Claude API.

#### Scenario: Call Claude with valid request
- **WHEN** ClaudeAdapter receives a call request
- **THEN** it formats request per Anthropic API spec, sends HTTP POST to api.anthropic.com, and parses response

#### Scenario: Handle Claude rate limit
- **WHEN** Claude API returns 429 status
- **THEN** adapter returns retriable error with retry-after duration

#### Scenario: Handle Claude authentication error
- **WHEN** Claude API returns 401 status
- **THEN** adapter returns non-retriable error "Invalid API key"

#### Scenario: Parse Claude token usage
- **WHEN** Claude API returns response with usage metadata
- **THEN** adapter extracts input_tokens and output_tokens and includes in CallResponse

### Requirement: System shall implement OpenAI adapter
The system SHALL implement an OpenAIAdapter that calls OpenAI GPT API.

#### Scenario: Call GPT with valid request
- **WHEN** OpenAIAdapter receives a call request
- **THEN** it formats request per OpenAI API spec, sends to api.openai.com, and parses response

#### Scenario: Handle OpenAI content filter
- **WHEN** OpenAI API returns content_filter error
- **THEN** adapter returns error "Content filtered by provider"

#### Scenario: Parse OpenAI token usage
- **WHEN** OpenAI API returns response with usage object
- **THEN** adapter extracts prompt_tokens and completion_tokens

### Requirement: System shall implement Ollama adapter
The system SHALL implement an OllamaAdapter that calls local Ollama API.

#### Scenario: Call Ollama with valid request
- **WHEN** OllamaAdapter receives a call request
- **THEN** it sends request to local Ollama endpoint (default localhost:11434)

#### Scenario: Handle Ollama model not found
- **WHEN** Ollama returns model not found error
- **THEN** adapter returns error "Model not available on Ollama server"

#### Scenario: Ollama cost estimation
- **WHEN** EstimateCost is called for Ollama model
- **THEN** adapter returns zero cost (local model)

### Requirement: Adapter shall normalize request parameters
The system SHALL normalize request parameters across different providers.

#### Scenario: Map temperature parameter
- **WHEN** request specifies temperature=0.7
- **THEN** adapter maps to provider-specific parameter (temperature for Claude/OpenAI, temp for Ollama)

#### Scenario: Map max_tokens parameter
- **WHEN** request specifies max_tokens=1000
- **THEN** adapter maps to provider-specific parameter (max_tokens for Claude/OpenAI, num_predict for Ollama)

#### Scenario: Handle unsupported parameter
- **WHEN** request includes parameter not supported by provider
- **THEN** adapter ignores the parameter and logs warning

### Requirement: Adapter shall normalize response format
The system SHALL normalize response format from different providers into unified CallResponse.

#### Scenario: Extract text from Claude response
- **WHEN** Claude returns response with content array
- **THEN** adapter extracts text from first content block

#### Scenario: Extract text from OpenAI response
- **WHEN** OpenAI returns response with choices array
- **THEN** adapter extracts message.content from first choice

#### Scenario: Extract text from Ollama response
- **WHEN** Ollama returns response with response field
- **THEN** adapter extracts text from response field

### Requirement: Adapter shall handle provider-specific errors
The system SHALL map provider-specific error codes to unified error types.

#### Scenario: Map authentication errors
- **WHEN** provider returns 401 or 403 status
- **THEN** adapter returns ErrAuthentication error type

#### Scenario: Map rate limit errors
- **WHEN** provider returns 429 status
- **THEN** adapter returns ErrRateLimit error type with retry_after

#### Scenario: Map quota exceeded errors
- **WHEN** provider returns quota exceeded error
- **THEN** adapter returns ErrQuotaExceeded error type

#### Scenario: Map timeout errors
- **WHEN** HTTP request times out
- **THEN** adapter returns ErrTimeout error type

### Requirement: Adapter shall implement request timeout
The system SHALL enforce timeout for all adapter calls.

#### Scenario: Call completes within timeout
- **WHEN** model responds within 60 seconds
- **THEN** adapter returns response normally

#### Scenario: Call exceeds timeout
- **WHEN** model does not respond within 60 seconds
- **THEN** adapter cancels request and returns ErrTimeout

### Requirement: Adapter shall support streaming responses
The system SHALL support streaming responses for compatible providers (Phase 2).

#### Scenario: Stream Claude response
- **WHEN** request enables streaming
- **THEN** ClaudeAdapter uses SSE endpoint and yields chunks as they arrive

#### Scenario: Provider does not support streaming
- **WHEN** request enables streaming for Ollama
- **THEN** adapter falls back to non-streaming call

### Requirement: Adapter shall calculate accurate costs
The system SHALL calculate costs based on actual token usage and provider pricing.

#### Scenario: Calculate Claude cost
- **WHEN** Claude call uses 1000 input tokens and 500 output tokens
- **THEN** adapter calculates cost as (1000 * input_price + 500 * output_price) / 1000000

#### Scenario: Calculate OpenAI cost
- **WHEN** OpenAI call uses tokens
- **THEN** adapter uses OpenAI pricing table for specified model

#### Scenario: Zero cost for local models
- **WHEN** Ollama call completes
- **THEN** adapter returns cost=0
