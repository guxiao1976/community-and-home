package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLostFoundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLostFoundLogic {
	return &GetLostFoundLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetLostFoundLogic) GetLostFound(req *types.GetLostFoundReq) (*types.GetLostFoundResp, error) {
	// 注入 JWT 身份 metadata，供 rpc 层 GetDataScopes 读过滤（T4.6 一致，评审 CRITICAL 补漏）
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.LostFoundServiceRpc.GetLostFound(callCtx, &communityv1.GetLostFoundRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	// 透出 rpc 业务错误（080006 数据权限拒绝 / 80004 不存在），与写接口一致
	if err := responsex.ToError(resp.GetBase()); err != nil {
		return nil, err
	}
	return &types.GetLostFoundResp{Item: toLostFoundItemInfo(resp.Item)}, nil
}
