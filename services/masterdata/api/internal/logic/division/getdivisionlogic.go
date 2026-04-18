package division

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionLogic {
	return &GetDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionLogic) GetDivision(req *types.GetDivisionReq) (resp *types.GetDivisionResp, err error) {
	d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区域不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区域详情失败")
	}

	if d.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("行政区域已删除")
	}

	return &types.GetDivisionResp{
		Division: toDivisionType(d),
	}, nil
}
