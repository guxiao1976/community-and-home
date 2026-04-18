package division

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDivisionLogic {
	return &DeleteDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDivisionLogic) DeleteDivision(req *types.DeleteDivisionReq) (resp *types.DeleteDivisionResp, err error) {
	// 1. Check existence
	existing, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("行政区域不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("行政区域已删除")
	}

	// 2. Check for active children
	count, err := l.svcCtx.MdAdministrativeDivisionModel.CountByParentId(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("查询子级行政区域失败")
	}
	if count > 0 {
		return nil, errorx.NewInvalidParamError("存在子级行政区域，无法删除")
	}

	// 3. Soft delete
	if err := l.svcCtx.MdAdministrativeDivisionModel.SoftDelete(l.ctx, req.Id); err != nil {
		return nil, errorx.NewDefaultError("删除行政区域失败")
	}

	return &types.DeleteDivisionResp{Success: true}, nil
}