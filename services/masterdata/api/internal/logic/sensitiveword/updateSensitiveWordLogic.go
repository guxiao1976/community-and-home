// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update sensitive word
func NewUpdateSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSensitiveWordLogic {
	return &UpdateSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSensitiveWordLogic) UpdateSensitiveWord(req *types.UpdateSensitiveWordReq) (resp *types.UpdateSensitiveWordResp, err error) {
	existing, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("敏感词不存在")
		}
		return nil, errorx.NewDefaultError("查询敏感词失败")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("敏感词已删除")
	}
	if existing.SubmissionStatus != 0 && existing.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("仅待提交或已拒绝状态的敏感词可以编辑")
	}

	// Capture change snapshot of current values before modification
	snapshot := map[string]interface{}{
		"word":     existing.Word,
		"category": existing.Category,
		"severity": existing.Severity,
		"action":   existing.Action,
		"status":   existing.Status,
	}
	snapshotJson, _ := json.Marshal(snapshot)
	existing.ChangeSnapshot = sql.NullString{String: string(snapshotJson), Valid: true}
	existing.SubmissionType = sql.NullInt64{Int64: 2, Valid: true}

	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Severity > 0 {
		if req.Severity < 1 || req.Severity > 3 {
			return nil, errorx.NewInvalidParamError("严重程度必须在1-3之间")
		}
		existing.Severity = int64(req.Severity)
	}
	if req.Action > 0 {
		if req.Action < 1 || req.Action > 3 {
			return nil, errorx.NewInvalidParamError("处理动作必须在1-3之间")
		}
		existing.Action = int64(req.Action)
	}
	if req.Status > 0 {
		if req.Status < 1 || req.Status > 2 {
			return nil, errorx.NewInvalidParamError("状态必须为1(启用)或2(禁用)")
		}
		existing.Status = int64(req.Status)
	}

	existing.SubmissionStatus = 0
	existing.UpdatedTime = time.Now()

	if err := l.svcCtx.MdSensitiveWordModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("更新敏感词失败")
	}

	return &types.UpdateSensitiveWordResp{Success: true}, nil
}
