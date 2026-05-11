package engine

import (
	"context"
	"fmt"

	"github.com/guxiao/community-and-home/services/moderation/internal/llm"
	"github.com/guxiao/community-and-home/services/moderation/internal/imagehash"
	"github.com/zeromicro/go-zero/core/logx"
)

type ImageEngine struct {
	hasher     *imagehash.ImageHasher
	smallModel llm.LLMClient
	largeModel llm.LLMClient
}

func NewImageEngine(
	h *imagehash.ImageHasher,
	sm, lm llm.LLMClient,
) *ImageEngine {
	return &ImageEngine{
		hasher:     h,
		smallModel: sm,
		largeModel: lm,
	}
}

func (e *ImageEngine) Check(ctx context.Context, imageData []byte, imgCtx string) (*ModerationResult, error) {
	if len(imageData) == 0 {
		return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
	}

	// 1. Perceptual hash check
	if e.hasher != nil {
		hash, err := e.hasher.Hash(nil)
		if err == nil && hash != 0 {
			_, category, dist, found := e.hasher.FindSimilar(hash)
			if found {
				return &ModerationResult{
					Pass:      false,
					RiskLevel: "high",
					Reason:    fmt.Sprintf("命中违规图片库 (距离=%d, 类别=%s)", dist, category),
					Details: []MatchDetail{{
						Layer:      "image_hash",
						Category:   category,
						Confidence: 1.0 - float64(dist)/64.0,
					}},
				}, nil
			}
		}
	}

	// 2. Small model
	if e.smallModel != nil {
		result, err := e.smallModel.CheckImage(ctx, imageData, imgCtx)
		if err == nil && result.Compliant && result.Confidence >= 0.9 {
			return &ModerationResult{Pass: true, RiskLevel: "low"}, nil
		}
		if err != nil {
			logx.Infof("small image model error: %v", err)
		}
	}

	// 3. Large model
	if e.largeModel == nil {
		return &ModerationResult{
			Pass:       true,
			RiskLevel:  "medium",
			NeedReview: true,
			Reason:     "模型不可用，需人工复审",
		}, nil
	}

	result, err := e.largeModel.CheckImage(ctx, imageData, imgCtx)
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
