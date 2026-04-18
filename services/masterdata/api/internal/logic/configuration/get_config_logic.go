package configuration

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigLogic {
	return &GetConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigLogic) GetConfig(req *types.GetConfigReq) (resp *types.GetConfigResp, err error) {
	config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("configuration not found")
	}

	return &types.GetConfigResp{
		Configuration: types.Configuration{
			Id:             config.Id,
			Module:         config.Module,
			ConfigKey:      config.ConfigKey,
			ConfigValue:    config.ConfigValue,
			ValueType:      config.ValueType,
			Description:    config.Description.String,
			IsPublic:       config.IsPublic,
			ApprovalStatus: config.ApprovalStatus,
			CreatedTime:    config.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:    config.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
