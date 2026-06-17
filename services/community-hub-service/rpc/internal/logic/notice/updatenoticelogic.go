package notice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeLogic {
	return &UpdateNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNoticeLogic) UpdateNotice(in *communityv1.UpdateNoticeRequest) (*communityv1.UpdateNoticeResponse, error) {
	// 校验存在
	notice, err := l.svcCtx.NoticeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("UpdateNotice: not found id=%d", in.Id)
		return &communityv1.UpdateNoticeResponse{
			Base: responsex.NewBaseRespWithError(80001, "通知不存在"),
		}, nil
	}

	isPinned := int32(0)
	if in.IsPinned {
		isPinned = 1
	}

	if err := l.svcCtx.NoticeModel.Update(l.ctx, in.Id, in.Title, in.Content, isPinned); err != nil {
		l.Errorf("UpdateNotice: update failed: %v", err)
		return nil, err
	}

	// Submit to moderation pipeline on content edit (async via Redis)
	publisherId := int64(0)
	if notice.PublisherId != nil {
		publisherId = *notice.PublisherId
	}

	contentSummary := in.Title
	if len([]rune(contentSummary)) > 100 {
		contentSummary = string([]rune(contentSummary)[:100])
	}

	auditResp, err := l.svcCtx.ModerationClient.CreateAuditLog(l.ctx, &moderationv1.CreateAuditLogRequest{
		ContentType:    "text",
		ContentSummary: contentSummary,
		RiskLevel:      "low",
		Pass:           false,
		SourceType:     "notice",
		SourceId:       in.Id,
		UserId:         publisherId,
	})
	if err != nil {
		l.Errorf("CreateAuditLog failed: %v", err)
		// Don't block the update — audit failure means human fallback later
	} else {
		textContent := in.Title + "\n" + in.Content
		taskMsg := map[string]interface{}{
			"task_id":      fmt.Sprintf("notice_%d_%d", in.Id, time.Now().UnixNano()),
			"audit_log_id": auditResp.Id,
			"source_type":  "notice",
			"source_id":    in.Id,
			"action":       "update",
			"items": []map[string]string{
				{"type": "text", "content": textContent, "field": "content"},
			},
		}
		body, _ := json.Marshal(taskMsg)
		if _, err := l.svcCtx.RedisClient.LpushCtx(l.ctx, "moderation:task:queue", string(body)); err != nil {
			l.Errorf("enqueue moderation task failed: %v", err)
		}
	}

	l.Infof("UpdateNotice success: id=%d", in.Id)
	return &communityv1.UpdateNoticeResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
