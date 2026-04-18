package configuration

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateConfigLogic {
	return &CreateConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateConfigLogic) CreateConfig(req *types.CreateConfigReq) (resp *types.CreateConfigResp, err error) {
	// Check if config key already exists in the same module
	existing, err := l.svcCtx.MdConfigurationModel.FindOneByModuleConfigKey(l.ctx, req.Module, req.ConfigKey)
	if err == nil && existing != nil {
		return nil, errorx.NewDefaultError("config key already exists in this module")
	}

	config := &model.MdConfiguration{
		Module:         req.Module,
		ConfigKey:      req.ConfigKey,
		ConfigValue:    req.ConfigValue,
		ValueType:      req.ValueType,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		IsPublic:       req.IsPublic,
		ApprovalStatus: 0,
		CreatedBy:      0, // TODO: Get from context
	}

	result, err := l.svcCtx.MdConfigurationModel.Insert(l.ctx, config)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to create configuration")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errorx.NewDefaultError("failed to get insert id")
	}

	return &types.CreateConfigResp{
		Id: id,
	}, nil
}
