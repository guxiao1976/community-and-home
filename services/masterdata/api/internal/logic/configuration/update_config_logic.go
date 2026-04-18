package configuration

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateConfigLogic {
	return &UpdateConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateConfigLogic) UpdateConfig(req *types.UpdateConfigReq) (resp *types.UpdateConfigResp, err error) {
	config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("configuration not found")
	}

	if req.ConfigValue != "" {
		config.ConfigValue = req.ConfigValue
	}
	if req.Description != "" {
		config.Description = sql.NullString{String: req.Description, Valid: true}
	}

	err = l.svcCtx.MdConfigurationModel.Update(l.ctx, config)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to update configuration")
	}

	return &types.UpdateConfigResp{
		Success: true,
	}, nil
}
