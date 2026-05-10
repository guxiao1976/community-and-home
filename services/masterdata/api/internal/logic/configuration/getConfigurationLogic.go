// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get single configuration details
func NewGetConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigurationLogic {
	return &GetConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigurationLogic) GetConfiguration(req *types.GetConfigurationReq) (resp *types.GetConfigurationResp, err error) {
	config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("配置不存在")
		}
		return nil, errorx.NewDefaultError("查询配置失败")
	}

	desc := ""
	if config.Description.Valid {
		desc = config.Description.String
	}

	return &types.GetConfigurationResp{
		Configuration: types.Configuration{
			Id:             config.Id,
			Module:         config.Module,
			Key:            config.ConfigKey,
			Value:          config.ConfigValue,
			ValueType:      config.ValueType,
			Description:    desc,
			IsPublic:       int32(config.IsPublic),
			ApprovalStatus: int32(config.ApprovalStatus),
			CreatedBy:      config.CreatedBy,
			CreatedTime:    config.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:    config.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
