package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLostFoundLogic {
	return &ResolveLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveLostFoundLogic) ResolveLostFound(in *communityv1.ResolveLostFoundRequest) (*communityv1.ResolveLostFoundResponse, error) {
	// 校验存在，并 reverse-lookup 内容 community_id（作为数据权限 target）
	item, err := l.svcCtx.LostFoundItemModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("ResolveLostFound: not found id=%d", in.Id)
		return &communityv1.ResolveLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80004, "寻失记录不存在"),
		}, nil
	}

	// 数据权限（T4.2）：落库前 AssertPublishScope(目标小区)。
	// 身份取调用方 JWT（gRPC metadata 注入）；本请求无 publisher_id 兜底，缺身份 → fail-closed 拒绝。
	denyResp, err := scope.CheckPublishScope(l.ctx, l.svcCtx.PermissionClient, item.CommunityId)
	if err != nil {
		l.Errorf("ResolveLostFound: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.ResolveLostFoundResponse{Base: denyResp}, nil
	}

	if err := l.svcCtx.LostFoundItemModel.UpdateStatus(l.ctx, in.Id, "resolved"); err != nil {
		l.Errorf("ResolveLostFound: update status failed: %v", err)
		return nil, err
	}

	l.Infof("ResolveLostFound success: id=%d", in.Id)
	return &communityv1.ResolveLostFoundResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
