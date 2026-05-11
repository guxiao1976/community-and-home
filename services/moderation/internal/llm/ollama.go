package llm

import (
	"context"
)

// OllamaClient is a local Ollama-based LLM moderation client.
// Methods are stubbed and return ErrNotImplemented until the Ollama
// integration is wired up.
type OllamaClient struct{}

// NewOllamaClient creates a new OllamaClient.
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{}
}

// CheckText sends text content to a local Ollama model for compliance checking.
func (c *OllamaClient) CheckText(_ context.Context, _, _ string) (*CheckResult, error) {
	return nil, ErrNotImplemented
}

// CheckImage sends image data to a local Ollama model for compliance checking.
func (c *OllamaClient) CheckImage(_ context.Context, _ []byte, _ string) (*CheckResult, error) {
	return nil, ErrNotImplemented
}
