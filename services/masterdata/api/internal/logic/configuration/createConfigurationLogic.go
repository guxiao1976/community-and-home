// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create new configuration
func NewCreateConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateConfigurationLogic {
	return &CreateConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateConfigurationLogic) CreateConfiguration(req *types.CreateConfigurationReq) (resp *types.CreateConfigurationResp, err error) {
	// 1. Validate value_type
	validTypes := map[string]bool{"string": true, "number": true, "boolean": true, "json": true}
	if !validTypes[req.ValueType] {
		return nil, errorx.NewInvalidParamError("无效的值类型，必须是 string、number、boolean 或 json")
	}

	// 2. Check if configuration already exists
	existing, err := l.svcCtx.MdConfigurationModel.FindOneByModuleConfigKey(l.ctx, req.Module, req.Key)
	if err == nil && existing != nil {
		return nil, errorx.NewInvalidParamError("配置键已存在")
	}

	// 3. Create configuration
	var description sql.NullString
	if req.Description != "" {
		description = sql.NullString{String: req.Description, Valid: true}
	}

	data := &model.MdConfiguration{
		Module:           req.Module,
		ConfigKey:        req.Key,
		ConfigValue:      req.Value,
		ValueType:        req.ValueType,
		Description:      description,
		IsPublic:         int64(req.IsPublic),
		ApprovalStatus:   0,
		SubmissionType:    sql.NullInt64{Int64: 1, Valid: true},
		SubmissionStatus: 0,
		CreatedBy:        0, // TODO: Get from JWT context
		CreatedTime:      time.Now(),
		UpdatedTime:      time.Now(),
	}

	res, err := l.svcCtx.MdConfigurationModel.Insert(l.ctx, data)
	if err != nil {
		return nil, errorx.NewDefaultError("创建配置失败")
	}

	id, _ := res.LastInsertId()
	return &types.CreateConfigurationResp{Id: id}, nil
}
