package text_review

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/types"
	"github.com/guxiao/community-and-home/services/moderation/internal/auditlog"
	"github.com/guxiao/community-and-home/services/moderation/internal/engine"

	"github.com/zeromicro/go-zero/core/logx"
)

type TextCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTextCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TextCheckLogic {
	return &TextCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TextCheckLogic) TextCheck(req *types.TextCheckReq) (*types.TextCheckResp, error) {
	if req.Content == "" {
		return nil, errorx.NewInvalidParamError("content is required")
	}
	if len(req.Content) > 10000 {
		return nil, errorx.NewInvalidParamError("content too long (max 10000)")
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "post"
	}

	checkMode := req.CheckMode
	if checkMode == "" {
		checkMode = "combined"
	}

	result, err := l.svcCtx.TextEngine.Check(l.ctx, req.Content, contentType, checkMode)
	if err != nil {
		logx.Errorf("text check error: %v", err)
		return nil, errorx.NewDefaultError("审核失败")
	}

	// Log audit record
	go l.logAudit(req, result)

	return &types.TextCheckResp{
		Pass:       result.Pass,
		RiskLevel:  result.RiskLevel,
		Reason:     result.Reason,
		NeedReview: result.NeedReview,
		Details:    convertDetails(result.Details),
	}, nil
}

func convertDetails(details []engine.MatchDetail) []types.MatchDetail {
	if details == nil {
		return []types.MatchDetail{}
	}
	out := make([]types.MatchDetail, 0, len(details))
	for _, d := range details {
		out = append(out, types.MatchDetail{
			Layer:       d.Layer,
			MatchedText: d.MatchedText,
			Category:    d.Category,
			Severity:    d.Severity,
			Confidence:  d.Confidence,
		})
	}
	return out
}

func (l *TextCheckLogic) logAudit(req *types.TextCheckReq, result *engine.ModerationResult) {
	contentType := req.ContentType
	if contentType == "" {
		contentType = "post"
	}

	// Truncate content for summary
	summary := req.Content
	if len(summary) > 100 {
		summary = summary[:100] + "..."
	}

	// Convert matched details
	matchedItems := make([]auditlog.MatchedItem, 0, len(result.Details))
	for _, d := range result.Details {
		matchedItems = append(matchedItems, auditlog.MatchedItem{
			Layer:       d.Layer,
			MatchedText: d.MatchedText,
			Category:    d.Category,
			Severity:    d.Severity,
			Confidence:  d.Confidence,
		})
	}

	// Determine check layer
	checkLayer := "ac_engine"
	if len(result.Details) > 0 {
		checkLayer = result.Details[0].Layer
	}

	entry := auditlog.LogEntry{
		ContentType:    contentType,
		ContentSummary: summary,
		RiskLevel:      result.RiskLevel,
		Pass:           result.Pass,
		Reason:         result.Reason,
		CheckLayer:     checkLayer,
		MatchedItems:   matchedItems,
		NeedReview:     result.NeedReview,
	}

	if err := l.svcCtx.AuditLogger.Log(context.Background(), entry); err != nil {
		logx.Errorf("failed to log audit: %v", err)
	}
}
