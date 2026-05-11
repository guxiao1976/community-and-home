package llm

import (
	"context"
)

// RemoteLLMClient is an HTTP-based LLM moderation client that talks to a
// remote moderation service. Methods are stubbed and return ErrNotImplemented
// until the remote integration is wired up.
type RemoteLLMClient struct{}

// NewRemoteLLMClient creates a new RemoteLLMClient.
func NewRemoteLLMClient() *RemoteLLMClient {
	return &RemoteLLMClient{}
}

// CheckText sends text content to a remote LLM service for compliance checking.
func (c *RemoteLLMClient) CheckText(_ context.Context, _, _ string) (*CheckResult, error) {
	return nil, ErrNotImplemented
}

// CheckImage sends image data to a remote LLM service for compliance checking.
func (c *RemoteLLMClient) CheckImage(_ context.Context, _ []byte, _ string) (*CheckResult, error) {
	return nil, ErrNotImplemented
}
