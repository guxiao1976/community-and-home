// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

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

type UpdateConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update configuration
func NewUpdateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigurationLogic {
	return &UpdateConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateConfigurationLogic) UpdateConfiguration(req *types.UpdateConfigurationReq) (resp *types.UpdateConfigurationResp, err error) {
	existing, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("配置不存在")
		}
		return nil, errorx.NewDefaultError("查询配置失败")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("配置已删除")
	}
	if existing.SubmissionStatus != 0 && existing.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("仅待提交或已拒绝状态的配置可以编辑")
	}

	// Capture change snapshot of current values before modification
	snapshot := map[string]interface{}{
		"config_value": existing.ConfigValue,
		"is_public":    existing.IsPublic,
	}
	if existing.Description.Valid {
		snapshot["description"] = existing.Description.String
	}
	snapshotJson, _ := json.Marshal(snapshot)
	existing.ChangeSnapshot = sql.NullString{String: string(snapshotJson), Valid: true}
	existing.SubmissionType = sql.NullInt64{Int64: 2, Valid: true}

	if req.Value != "" {
		existing.ConfigValue = req.Value
	}
	if req.Description != "" {
		existing.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.IsPublic > 0 {
		existing.IsPublic = int64(req.IsPublic)
	}

	existing.SubmissionStatus = 0
	existing.UpdatedTime = time.Now()

	if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("更新配置失败")
	}

	return &types.UpdateConfigurationResp{Success: true}, nil
}
