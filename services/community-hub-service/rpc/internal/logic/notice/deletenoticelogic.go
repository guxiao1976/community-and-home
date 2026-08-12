package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNoticeLogic {
	return &DeleteNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteNoticeLogic) DeleteNotice(in *communityv1.DeleteNoticeRequest) (*communityv1.DeleteNoticeResponse, error) {
	// 校验存在，并 reverse-lookup 内容 community_id（作为数据权限 target）
	notice, err := l.svcCtx.NoticeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("DeleteNotice: not found id=%d", in.Id)
		return &communityv1.DeleteNoticeResponse{
			Base: responsex.NewBaseRespWithError(80001, "通知不存在"),
		}, nil
	}

	// 数据权限（T4.3）：删除前 AssertPublishScope(目标 community_id)，拒绝则不删。
	denyResp, err := scope.CheckPublishScope(l.ctx, l.svcCtx.PermissionClient, notice.CommunityId)
	if err != nil {
		l.Errorf("DeleteNotice: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.DeleteNoticeResponse{Base: denyResp}, nil
	}

	if err := l.svcCtx.NoticeModel.SoftDelete(l.ctx, in.Id); err != nil {
		l.Errorf("DeleteNotice: soft delete failed: %v", err)
		return nil, err
	}

	l.Infof("DeleteNotice success: id=%d", in.Id)
	return &communityv1.DeleteNoticeResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
