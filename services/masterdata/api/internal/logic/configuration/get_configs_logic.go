package configuration

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigsLogic {
	return &GetConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigsLogic) GetConfigs(req *types.GetConfigsReq) (resp *types.GetConfigsResp, err error) {
	offset := (req.Page - 1) * req.PageSize

	var configs []*model.MdConfiguration
	var total int64

	if req.Module != "" {
		configs, total, err = l.svcCtx.MdConfigurationModel.FindByModule(l.ctx, req.Module, int(req.PageSize), int(offset))
	} else {
		configs, total, err = l.svcCtx.MdConfigurationModel.FindAll(l.ctx, int(req.PageSize), int(offset))
	}

	if err != nil {
		return nil, err
	}

	list := make([]types.Configuration, 0, len(configs))
	for _, config := range configs {
		list = append(list, types.Configuration{
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
		})
	}

	return &types.GetConfigsResp{
		List:  list,
		Total: total,
	}, nil
}
