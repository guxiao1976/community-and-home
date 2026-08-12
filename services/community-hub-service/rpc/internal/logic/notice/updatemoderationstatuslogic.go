package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeModerationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNoticeModerationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeModerationStatusLogic {
	return &UpdateNoticeModerationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateNoticeModerationStatus 审核状态回调（moderation-service 服务间调用，无用户 JWT）。
//
// 评审 S4（T4.5）落地：
//  1. reverse-lookup 内容 id → community_id（查不到 → 拒绝，目标小区被校验而非假设）；
//  2. 以系统身份（system_user_id=0，global scope）调 AssertPublishScope（服务身份回调放行，不按作者 scope）。
//
// SEE: [[is-system-no-permission-shortcut]] — 系统身份走 grant 判定，无字段短路
func (l *UpdateNoticeModerationStatusLogic) UpdateNoticeModerationStatus(in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	// reverse-lookup 内容 community_id
	notice, err := l.svcCtx.NoticeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("UpdateNoticeModerationStatus: content not found id=%d", in.Id)
		return &communityv1.UpdateModerationStatusResponse{
			Base: responsex.NewBaseRespWithError(80001, "通知不存在"),
		}, nil
	}

	// 系统身份数据权限校验（global 放行；目标小区不存在/未知 → 安全拒绝）
	denyResp, err := scope.CheckSystemPublishScope(l.ctx, l.svcCtx.PermissionClient, notice.CommunityId)
	if err != nil {
		l.Errorf("UpdateNoticeModerationStatus: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.UpdateModerationStatusResponse{Base: denyResp}, nil
	}

	if err := l.svcCtx.NoticeModel.UpdateModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
		l.Errorf("UpdateNoticeModerationStatus failed: %v", err)
		return nil, err
	}
	l.Infof("UpdateNoticeModerationStatus: id=%d, status=%d", in.Id, in.ModerationStatus)
	return &communityv1.UpdateModerationStatusResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
