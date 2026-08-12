package notice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNoticeLogic {
	return &CreateNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNoticeLogic) CreateNotice(in *communityv1.CreateNoticeRequest) (*communityv1.CreateNoticeResponse, error) {
	if in.Title == "" || in.Content == "" {
		return &communityv1.CreateNoticeResponse{
			Base: responsex.NewBaseRespWithError(80005, "标题和内容不能为空"),
		}, nil
	}

	// 数据权限（T4.3）：落库前 AssertPublishScope(目标 community_id)。
	denyResp, err := scope.CheckPublishScope(l.ctx, l.svcCtx.PermissionClient, in.GetCommunityId())
	if err != nil {
		l.Errorf("CreateNotice: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.CreateNoticeResponse{Base: denyResp}, nil
	}

	noticeRole := roleToString(in.Role)
	id := snowflake.NextID()

	n := &model.Notice{
		Id:          id,
		CommunityId: in.CommunityId,
		Title:       in.Title,
		Content:     in.Content,
		Role:        noticeRole,
		Publisher:   in.Publisher,
		PublisherId: &in.PublisherId,
		IsPinned:    0,
		PublishedAt: time.Now(),
	}

	if _, err := l.svcCtx.NoticeModel.Insert(l.ctx, n); err != nil {
		l.Errorf("CreateNotice: insert failed: %v", err)
		return nil, err
	}

	// 5. Submit to moderation pipeline (async via Redis)
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
		SourceId:       id,
		UserId:         in.PublisherId,
	})
	if err != nil {
		l.Errorf("CreateAuditLog failed: %v", err)
		// Don't block the publish — audit failure means human fallback later
	} else {
		// Build task message with full content
		textContent := in.Title + "\n" + in.Content
		taskMsg := map[string]interface{}{
			"task_id":      fmt.Sprintf("notice_%d_%d", id, time.Now().UnixNano()),
			"audit_log_id": auditResp.Id,
			"source_type":  "notice",
			"source_id":    id,
			"action":       "create",
			"items": []map[string]string{
				{"type": "text", "content": textContent, "field": "content"},
			},
		}
		body, _ := json.Marshal(taskMsg)
		if _, err := l.svcCtx.RedisClient.LpushCtx(l.ctx, "moderation:task:queue", string(body)); err != nil {
			l.Errorf("enqueue moderation task failed: %v", err)
		}
	}

	l.Infof("CreateNotice success: id=%d, communityId=%d", id, in.CommunityId)
	return &communityv1.CreateNoticeResponse{
		Base: responsex.NewBaseResp(),
		Id:   id,
	}, nil
}
