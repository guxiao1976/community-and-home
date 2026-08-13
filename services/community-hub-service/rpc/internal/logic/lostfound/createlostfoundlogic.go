package lostfound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLostFoundLogic {
	return &CreateLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLostFoundLogic) CreateLostFound(in *communityv1.CreateLostFoundRequest) (*communityv1.CreateLostFoundResponse, error) {
	if in.Title == "" {
		return &communityv1.CreateLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80005, "标题不能为空"),
		}, nil
	}

	// 数据权限（T4.2）：落库前 AssertPublishScope，校验顺序：功能权限 → 数据权限 → 落库。
	// 身份仅取调用方 JWT（API 层经 gRPC metadata 注入），不信任 body publisher_id（防伪造）。
	denyResp, err := scope.CheckPublishScope(l.ctx, l.svcCtx.PermissionClient, in.GetCommunityId())
	if err != nil {
		l.Errorf("CreateLostFound: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.CreateLostFoundResponse{Base: denyResp}, nil
	}

	// 板块发布配额（Task 4.3 / design §4.3）：AssertPublishScope 之后、落库之前校验。
	// 口径「用户×小区×板块」，按目标小区计；超限 → 080007；GetSectionQuota/计数传输错误原样传播（fail-closed）。
	if err := scope.CheckSectionQuota(l.ctx, l.svcCtx, in.PublisherId, in.GetCommunityId(), scope.SectionTypeLostFound); err != nil {
		if ce := errx.FromError(err); ce != nil && ce.Code == scope.CodeSectionQuotaExceeded {
			return &communityv1.CreateLostFoundResponse{Base: responsex.NewBaseRespWithError(scope.CodeSectionQuotaExceeded, "超出发布配额")}, nil
		}
		l.Errorf("CreateLostFound: check section quota failed: %v", err)
		return nil, err
	}

	imageUrlsJSON := "[]"
	if len(in.ImageUrls) > 0 {
		b, _ := json.Marshal(in.ImageUrls)
		imageUrlsJSON = string(b)
	}

	id := snowflake.NextID()
	item := &model.LostFoundItem{
		Id:           id,
		CommunityId:  in.CommunityId,
		Type:         typeToString(in.Type),
		Title:        in.Title,
		Description:  in.Description,
		ImageUrls:    imageUrlsJSON,
		ContactPhone: in.ContactPhone,
		Status:       "active",
		PublisherId:  in.PublisherId,
	}

	if _, err := l.svcCtx.LostFoundItemModel.Insert(l.ctx, item); err != nil {
		l.Errorf("CreateLostFound: insert failed: %v", err)
		return nil, err
	}

	// Build text content for moderation
	textContent := in.Title
	if in.Description != "" {
		textContent += "\n" + in.Description
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
		SourceType:     "lost_found",
		SourceId:       id,
		UserId:         in.PublisherId,
	})
	if err != nil {
		l.Errorf("CreateAuditLog failed: %v", err)
	} else {
		items := []map[string]string{
			{"type": "text", "content": textContent, "field": "content"},
		}
		// Image audit reservation
		for _, imgUrl := range in.ImageUrls {
			items = append(items, map[string]string{"type": "image", "content": imgUrl, "field": "image"})
		}
		taskMsg := map[string]interface{}{
			"task_id":      fmt.Sprintf("lost_found_%d_%d", id, time.Now().UnixNano()),
			"audit_log_id": auditResp.Id,
			"source_type":  "lost_found",
			"source_id":    id,
			"action":       "create",
			"items":        items,
		}
		body, _ := json.Marshal(taskMsg)
		if _, err := l.svcCtx.RedisClient.LpushCtx(l.ctx, "moderation:task:queue", string(body)); err != nil {
			l.Errorf("enqueue moderation task failed: %v", err)
		}
	}

	l.Infof("CreateLostFound success: id=%d, communityId=%d", id, in.CommunityId)
	return &communityv1.CreateLostFoundResponse{
		Base: responsex.NewBaseResp(),
		Id:   id,
	}, nil
}
