// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const errHasChildren = "该区划下存在下级数据，无法删除"

func checkChildData(ctx context.Context, svcCtx *svc.ServiceContext, divId int64, divLevel int64) error {
	childCount, err := svcCtx.MdAdministrativeDivisionModel.CountByParentId(ctx, divId)
	if err != nil {
		return errorx.NewDefaultError("检查下级区划失败")
	}
	if childCount > 0 {
		return errorx.NewDefaultError(errHasChildren)
	}

	var countyId, streetId, communityDivId *int64
	switch divLevel {
	case 3:
		countyId = &divId
	case 4:
		streetId = &divId
	case 5:
		communityDivId = &divId
	default:
		return nil
	}

	areaCount, err := svcCtx.MdResidentialAreaModel.Count(ctx, countyId, streetId, communityDivId, nil, nil, nil, nil)
	if err != nil {
		return errorx.NewDefaultError("检查关联小区失败")
	}
	if areaCount > 0 {
		return errorx.NewDefaultError(errHasChildren)
	}
	return nil
}

type DeleteDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Soft delete administrative division
func NewDeleteDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDivisionLogic {
	return &DeleteDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDivisionLogic) DeleteDivision(req *types.DeleteDivisionReq) (resp *types.DeleteDivisionResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区划不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区划失败")
	}
	if div.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("行政区划已删除")
	}
	if div.SubmissionStatus != 0 || !div.SubmissionType.Valid || div.SubmissionType.Int64 != 1 {
		return nil, errorx.NewDefaultError("仅新建待提交的数据可以物理删除，已批准数据请使用发起删除功能")
	}

	if err := checkChildData(l.ctx, l.svcCtx, req.Id, div.Level); err != nil {
		return nil, err
	}

	if err := l.svcCtx.MdAdministrativeDivisionModel.Delete(l.ctx, req.Id); err != nil {
		return nil, errorx.NewDefaultError("删除行政区划失败: " + err.Error())
	}

	return &types.DeleteDivisionResp{Success: true}, nil
}
