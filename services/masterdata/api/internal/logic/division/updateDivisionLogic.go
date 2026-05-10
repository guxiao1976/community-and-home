// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update administrative division
func NewUpdateDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDivisionLogic {
	return &UpdateDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDivisionLogic) UpdateDivision(req *types.UpdateDivisionReq) (resp *types.UpdateDivisionResp, err error) {
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
	if div.SubmissionStatus != 0 && div.SubmissionStatus != 2 && div.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("仅待提交、已批准或已拒绝状态的区划可以编辑")
	}

	// If editing an approved record, capture snapshot and set type=2 (modify pending)
	if div.SubmissionStatus == 2 {
		snapshot := map[string]interface{}{
			"name":       div.Name,
			"code":       div.Code,
			"sort_order": div.SortOrder,
			"status":     div.Status,
		}
		if div.ParentId.Valid {
			snapshot["parent_id"] = div.ParentId.Int64
		}
		snapshotJson, _ := json.Marshal(snapshot)
		div.ChangeSnapshot = sql.NullString{String: string(snapshotJson), Valid: true}
		div.SubmissionType = sql.NullInt64{Int64: 2, Valid: true}
		div.SubmissionStatus = 0
	}

	if req.Name != "" {
		div.Name = req.Name
	}
	if req.Code != "" {
		div.Code = req.Code
	}
	if req.SortOrder > 0 {
		div.SortOrder = int64(req.SortOrder)
	}
	if req.Status > 0 {
		div.Status = int64(req.Status)
	}

	if err := l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, div); err != nil {
		return nil, errorx.NewDefaultError("更新行政区划失败: " + err.Error())
	}

	return &types.UpdateDivisionResp{Success: true}, nil
}
