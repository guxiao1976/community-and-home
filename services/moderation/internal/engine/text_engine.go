package engine

import (
	"context"

	"github.com/guxiao/community-and-home/services/moderation/internal/ac"
	"github.com/guxiao/community-and-home/services/moderation/internal/llm"
	"github.com/guxiao/community-and-home/services/moderation/internal/normalize"
	"github.com/guxiao/community-and-home/services/moderation/internal/splitword"
	"github.com/guxiao/community-and-home/services/moderation/internal/whitelist"
	"github.com/zeromicro/go-zero/core/logx"
)

type TextEngine struct {
	normalizer  *normalize.Normalizer
	acMachine   *ac.ACMachine
	wl          *whitelist.Whitelist
	splitter    *splitword.SplitDetector
	smallModel  llm.LLMClient
	largeModel  llm.LLMClient
	highConfTh  float64
}

func NewTextEngine(
	n *normalize.Normalizer,
	acm *ac.ACMachine,
	wl *whitelist.Whitelist,
	sd *splitword.SplitDetector,
	sm, lm llm.LLMClient,
	highConfTh float64,
) *TextEngine {
	return &TextEngine{
		normalizer: n,
		acMachine:  acm,
		wl:         wl,
		splitter:   sd,
		smallModel: sm,
		largeModel: lm,
		highConfTh: highConfTh,
	}
}

func (e *TextEngine) Check(ctx context.Context, content, contentType string) (*ModerationResult, error) {
	if content == "" {
		return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
	}

	// 1. Normalize
	normalized, posMap := e.normalizer.Normalize(content)

	// 2. AC automaton match
	acHits := e.acMachine.Match(normalized)

	// 3. Whitelist filter: skip segments where whitelist >= blacklist
	if len(acHits) > 0 {
		_, wlLen := e.wl.LongestMatch(normalized)
		var filtered []MatchDetail
		for _, h := range acHits {
			if wlLen > 0 && h.End-h.Start <= wlLen {
				continue
			}
			detail := MatchDetail{
				Layer:       "ac_engine",
				MatchedText: h.Word,
				Category:    h.Category,
				Severity:    h.Severity,
				Confidence:  1.0,
			}
			filtered = append(filtered, detail)
		}
		acHits = nil // stop using acHits directly
		if len(filtered) == 0 {
			return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
		}

		// 4. Check severity: high (1) → direct reject
		for _, d := range filtered {
			if d.Severity == 1 {
				return &ModerationResult{
					Pass:      false,
					RiskLevel: "high",
					Reason:    "命中敏感词: " + d.MatchedText,
					Details:   filtered,
				}, nil
			}
		}

		// 5. Gray list (severity 2/3) → try small model
		err := e.trySmallModel(ctx, content, contentType, filtered)
		if err == nil {
			// small model returned a clear result
			_ = posMap
			return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
		}
		logx.Infof("small model not available or inconclusive: %v", err)
	}

	// 6. Split word detection
	splitHits := e.splitter.Detect(content, e.acMachine)
	if len(splitHits) > 0 {
		var details []MatchDetail
		for _, h := range splitHits {
			details = append(details, MatchDetail{
				Layer:       "ac_engine",
				MatchedText: h.Word,
				Category:    h.Category,
				Severity:    h.Severity,
				Confidence:  0.9,
			})
		}
		if details[0].Severity == 1 {
			return &ModerationResult{
				Pass:      false,
				RiskLevel: "high",
				Reason:    "命中拆字变体: " + details[0].MatchedText,
				Details:   details,
			}, nil
		}
		// gray list from split detection → model layer
		err := e.trySmallModel(ctx, content, contentType, details)
		if err == nil {
			return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
		}
		logx.Infof("small model not available for split detection: %v", err)
	}

	// 7. Large model (or fallback)
	return e.tryLargeModel(ctx, content, contentType)
}

func (e *TextEngine) trySmallModel(ctx context.Context, content, contentType string, details []MatchDetail) error {
	if e.smallModel == nil {
		return llm.ErrNotImplemented
	}
	result, err := e.smallModel.CheckText(ctx, content, contentType)
	if err != nil {
		return err
	}
	if result.Compliant && result.Confidence >= e.highConfTh {
		return nil
	}
	return err
}

func (e *TextEngine) tryLargeModel(ctx context.Context, content, contentType string) (*ModerationResult, error) {
	if e.largeModel == nil {
		// Both models unavailable → pass with review flag
		return &ModerationResult{
			Pass:       true,
			RiskLevel:  "medium",
			NeedReview: true,
			Reason:     "模型不可用，需人工复审",
		}, nil
	}

	result, err := e.largeModel.CheckText(ctx, content, contentType)
	if err != nil {
		return &ModerationResult{
			Pass:       true,
			RiskLevel:  "medium",
			NeedReview: true,
			Reason:     "大模型不可用，需人工复审",
		}, nil
	}

	if result.Compliant {
		return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
	}
	return &ModerationResult{
		Pass:      false,
		RiskLevel: "high",
		Reason:    result.Reason,
		NeedReview: true,
		Details: []MatchDetail{{
			Layer:      "large_model",
			Category:   result.Category,
			Confidence: result.Confidence,
		}},
	}, nil
}
