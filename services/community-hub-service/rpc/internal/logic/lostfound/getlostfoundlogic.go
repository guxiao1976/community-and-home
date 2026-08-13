package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLostFoundLogic {
	return &GetLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLostFoundLogic) GetLostFound(in *communityv1.GetLostFoundRequest) (*communityv1.GetLostFoundResponse, error) {
	// 审核可见性门禁：读路径仅返回审核通过（moderation_status=1）的内容，
	// 待审核/拒绝的内容对普通用户不可见（FindOnePublished 过滤后 ErrNoRows → 80004）。
	it, err := l.svcCtx.LostFoundItemModel.FindOnePublished(l.ctx, in.Id)
	if err != nil {
		l.Infof("GetLostFound: not found id=%d", in.Id)
		return &communityv1.GetLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80004, "寻失记录不存在"),
		}, nil
	}

	// 数据范围读过滤（评审 CRITICAL 补漏：T4.6 只覆盖 List，Get-by-ID 漏网）：
	// reverse-lookup 内容 community_id → FilterAllowed（GLOBAL 放行 / LIMITED IN / EMPTY+无身份拒绝）。
	// 防止 LIMITED/EMPTY 用户按 ID（Snowflake 时间有序可枚举/分享链接）越权读取 description/contact_phone。
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, it.CommunityId)
	if err != nil {
		l.Errorf("GetLostFound: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		l.Infof("GetLostFound: denied by data scope id=%d community=%d", in.Id, it.CommunityId)
		return &communityv1.GetLostFoundResponse{
			Base: scope.DenyBase(),
		}, nil
	}

	return &communityv1.GetLostFoundResponse{
		Base: responsex.NewBaseResp(),
		Item: toProtoLostFoundItem(it),
	}, nil
}
