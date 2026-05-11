package llm

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned when an LLM method has not been wired up yet.
var ErrNotImplemented = errors.New("llm client not implemented")

// CheckResult is the outcome of a single content-check call to an LLM backend.
type CheckResult struct {
	Compliant  bool    `json:"compliant"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	Category   string  `json:"category,omitempty"`
}

// LLMClient abstracts text and image content moderation via an LLM service.
type LLMClient interface {
	CheckText(ctx context.Context, content, contentType string) (*CheckResult, error)
	CheckImage(ctx context.Context, imageData []byte, imgCtx string) (*CheckResult, error)
}
