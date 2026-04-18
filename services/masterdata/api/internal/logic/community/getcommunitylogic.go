package community

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommunityLogic {
	return &GetCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommunityLogic) GetCommunity(req *types.GetCommunityReq) (resp *types.GetCommunityResp, err error) {
	c, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("社区不存在")
		}
		return nil, errorx.NewDefaultError("查询社区详情失败")
	}

	if c.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("社区已删除")
	}

	return &types.GetCommunityResp{
		Community: toCommunityType(c),
	}, nil
}