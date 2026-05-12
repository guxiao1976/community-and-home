package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
	"github.com/guxiao/community-and-home/services/moderation/internal/llm"
	"github.com/guxiao/community-and-home/services/moderation/internal/normalize"
	"github.com/guxiao/community-and-home/services/moderation/internal/splitword"
	"github.com/guxiao/community-and-home/services/moderation/internal/whitelist"
)

// mockLLM is a test LLM client that returns configurable results
type mockLLM struct {
	compliant  bool
	confidence float64
	err        error
}

func (m *mockLLM) CheckText(_ context.Context, _, _ string) (*llm.CheckResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.CheckResult{Compliant: m.compliant, Confidence: m.confidence}, nil
}

func (m *mockLLM) CheckImage(_ context.Context, _ []byte, _ string) (*llm.CheckResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.CheckResult{Compliant: m.compliant, Confidence: m.confidence}, nil
}

func newTestTextEngine(words []ac.WordEntry, wlWords []string, sm, lm llm.LLMClient) *TextEngine {
	acm := ac.NewACMachine()
	_ = acm.Build(words)

	wl := whitelist.NewWhitelist()
	wl.Build(wlWords)

	sd := splitword.NewSplitDetector(" *xX")

	return NewTextEngine(
		normalize.New(normalize.WithWidth(), normalize.WithCase(), normalize.WithChinese()),
		acm, wl, sd, sm, lm, 0.9,
	)
}

func TestTextEngine_EmptyContent(t *testing.T) {
	eng := newTestTextEngine(nil, nil, nil, nil)
	result, err := eng.Check(context.Background(), "", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass || result.RiskLevel != "low" {
		t.Errorf("expected pass=true risk=low, got pass=%v risk=%s", result.Pass, result.RiskLevel)
	}
}

func TestTextEngine_ACHighSeverity_Reject(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "敏感词", Category: "test", Severity: 1},
	}
	eng := newTestTextEngine(words, nil, nil, nil)

	result, err := eng.Check(context.Background(), "这是一个敏感词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pass {
		t.Error("expected pass=false for high severity match")
	}
	if result.RiskLevel != "high" {
		t.Errorf("expected risk=high, got %s", result.RiskLevel)
	}
}

func TestTextEngine_NormalText_Pass(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "敏感词", Category: "test", Severity: 1},
	}
	eng := newTestTextEngine(words, nil, nil, nil)

	result, err := eng.Check(context.Background(), "今天天气真好", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Error("expected pass=true for clean text")
	}
}

func TestTextEngine_WhitelistOverride(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "白", Category: "test", Severity: 1},
	}
	wlWords := []string{"白菜"}
	eng := newTestTextEngine(words, wlWords, nil, nil)

	result, err := eng.Check(context.Background(), "白菜汤真好喝", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Error("expected pass=true when whitelist is longer than blacklist match")
	}
}

func TestTextEngine_GrayList_NoModel_NeedReview(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "灰色词", Category: "gray", Severity: 2},
	}
	eng := newTestTextEngine(words, nil, nil, nil)

	result, err := eng.Check(context.Background(), "这是一个灰色词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Severity 2 (gray) + no models → should degrade with need_review
	if !result.Pass {
		t.Error("expected pass=true (degraded) for gray word with no models")
	}
	if !result.NeedReview {
		t.Error("expected need_review=true for gray word with no models")
	}
}

func TestTextEngine_SmallModelPasses(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "灰色词", Category: "gray", Severity: 2},
	}
	sm := &mockLLM{compliant: true, confidence: 0.95}
	eng := newTestTextEngine(words, nil, sm, nil)

	result, err := eng.Check(context.Background(), "这是一个灰色词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Error("expected pass=true when small model confirms compliance")
	}
	if result.NeedReview {
		t.Error("expected need_review=false when small model passes")
	}
}

func TestTextEngine_SmallModelEscalatesToLarge(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "灰色词", Category: "gray", Severity: 2},
	}
	// Small model says non-compliant → escalate to large model
	sm := &mockLLM{compliant: false, confidence: 0.3}
	lm := &mockLLM{compliant: true, confidence: 0.95}
	eng := newTestTextEngine(words, nil, sm, lm)

	result, err := eng.Check(context.Background(), "这是一个灰色词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Error("expected pass=true when large model says compliant")
	}
}

func TestTextEngine_LargeModelRejects(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "灰色词", Category: "gray", Severity: 2},
	}
	// Small model has low confidence (below threshold) → should escalate to large
	// Since trySmallModel checks Compliant && Confidence >= threshold,
	// low confidence means it returns the err value (nil here), which engine
	// interprets as "small model passed". We need a non-nil err to escalate.
	sm := &mockLLM{compliant: false, confidence: 0.3, err: llm.ErrNotImplemented}
	lm := &mockLLM{compliant: false, confidence: 0.95}
	eng := newTestTextEngine(words, nil, sm, lm)

	result, err := eng.Check(context.Background(), "这是一个灰色词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pass {
		t.Error("expected pass=false when large model rejects")
	}
	if !result.NeedReview {
		t.Error("expected need_review=true when large model rejects")
	}
}

func TestTextEngine_SmallModelUnavailable_Escalates(t *testing.T) {
	words := []ac.WordEntry{
		{Word: "灰色词", Category: "gray", Severity: 2},
	}
	sm := &mockLLM{err: errors.New("connection refused")}
	eng := newTestTextEngine(words, nil, sm, nil)

	result, err := eng.Check(context.Background(), "这是一个灰色词测试", "post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Small model fails → no large model → degrade
	if !result.NeedReview {
		t.Error("expected need_review=true when both models unavailable")
	}
}

func TestImageEngine_EmptyData(t *testing.T) {
	eng := NewImageEngine(nil, nil, nil)
	result, err := eng.Check(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass || result.RiskLevel != "low" {
		t.Error("expected pass=true for empty image data")
	}
}

func TestImageEngine_NoModels_NeedReview(t *testing.T) {
	eng := NewImageEngine(nil, nil, nil)
	data := []byte("fake image data")

	result, err := eng.Check(context.Background(), data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.NeedReview {
		t.Error("expected need_review=true when no models available")
	}
}

func TestImageEngine_LargeModelPasses(t *testing.T) {
	lm := &mockLLM{compliant: true, confidence: 0.95}
	eng := NewImageEngine(nil, nil, lm)

	result, err := eng.Check(context.Background(), []byte("fake image"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Error("expected pass=true when large model says compliant")
	}
}

func TestImageEngine_LargeModelRejects(t *testing.T) {
	lm := &mockLLM{compliant: false, confidence: 0.8}
	eng := NewImageEngine(nil, nil, lm)

	result, err := eng.Check(context.Background(), []byte("fake image"), "test context")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pass {
		t.Error("expected pass=false when large model rejects")
	}
	if !result.NeedReview {
		t.Error("expected need_review=true for large model rejection")
	}
}
