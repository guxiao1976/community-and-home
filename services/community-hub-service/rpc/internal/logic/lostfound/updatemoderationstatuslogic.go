package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLostFoundModerationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLostFoundModerationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLostFoundModerationStatusLogic {
	return &UpdateLostFoundModerationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateLostFoundModerationStatus 审核状态回调（moderation-service 服务间调用，无用户 JWT）。
//
// 评审 S4（T4.5）落地：reverse-lookup 内容 community_id → 系统身份（system_user_id=0）
// AssertPublishScope 校验（global 放行，不按作者 scope）；内容不存在 → 拒绝。
//
// SEE: [[is-system-no-permission-shortcut]] — 系统身份走 grant 判定，无字段短路
func (l *UpdateLostFoundModerationStatusLogic) UpdateLostFoundModerationStatus(in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	// reverse-lookup 内容 community_id
	item, err := l.svcCtx.LostFoundItemModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("UpdateLostFoundModerationStatus: content not found id=%d", in.Id)
		return &communityv1.UpdateModerationStatusResponse{
			Base: responsex.NewBaseRespWithError(80004, "寻失记录不存在"),
		}, nil
	}

	denyResp, err := scope.CheckSystemPublishScope(l.ctx, l.svcCtx.PermissionClient, item.CommunityId)
	if err != nil {
		l.Errorf("UpdateLostFoundModerationStatus: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.UpdateModerationStatusResponse{Base: denyResp}, nil
	}

	if err := l.svcCtx.LostFoundItemModel.UpdateModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
		l.Errorf("UpdateLostFoundModerationStatus failed: %v", err)
		return nil, err
	}
	l.Infof("UpdateLostFoundModerationStatus: id=%d, status=%d", in.Id, in.ModerationStatus)
	return &communityv1.UpdateModerationStatusResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
