package division

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDivisionLogic {
	return &UpdateDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDivisionLogic) UpdateDivision(req *types.UpdateDivisionReq) (resp *types.UpdateDivisionResp, err error) {
	// 1. Find existing
	existing, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("行政区域不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("行政区域已删除")
	}

	// 2. Handle parent change (recalculate path)
	if req.ParentId != nil {
		oldPath := existing.Path
		if *req.ParentId == 0 {
			// Moving to root level
			existing.ParentId = sql.NullInt64{Valid: false}
			existing.Level = 1
			existing.Path = "/"
		} else {
			parent, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, *req.ParentId)
			if err != nil {
				return nil, errorx.NewInvalidParamError("上级行政区域不存在")
			}
			if parent.DeleteTime.Valid {
				return nil, errorx.NewInvalidParamError("上级行政区域已删除")
			}
			existing.ParentId = sql.NullInt64{Int64: *req.ParentId, Valid: true}
			existing.Level = parent.Level + 1
			newPath := fmt.Sprintf("%s%d/", parent.Path, parent.Id)
			existing.Path = newPath

			// Update descendants paths
			if err := l.svcCtx.MdAdministrativeDivisionModel.UpdatePathForDescendants(l.ctx, oldPath, newPath); err != nil {
				return nil, errorx.NewDefaultError("更新子级路径失败")
			}
		}
	}

	// 3. Apply field updates
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.SortOrder != 0 {
		existing.SortOrder = int64(req.SortOrder)
	}
	if req.Status != 0 {
		existing.Status = int64(req.Status)
	}

	existing.UpdatedTime = time.Now()

	// 4. Save
	if err := l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("更新行政区域失败")
	}

	return &types.UpdateDivisionResp{Success: true}, nil
}