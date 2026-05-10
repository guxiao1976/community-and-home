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

type GetConfigurationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List configurations with filters
func NewGetConfigurationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigurationsLogic {
	return &GetConfigurationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConfigurationsLogic) GetConfigurations(req *types.GetConfigurationsReq) (resp *types.GetConfigurationsResp, err error) {
	// Calculate pagination
	limit := int(req.PageSize)
	offset := (int(req.Page) - 1) * limit

	var configs []*model.MdConfiguration
	var total int64

	// Query based on filters
	if req.Module != "" {
		configs, total, err = l.svcCtx.MdConfigurationModel.FindByModule(l.ctx, req.Module, limit, offset)
	} else {
		configs, total, err = l.svcCtx.MdConfigurationModel.FindAll(l.ctx, limit, offset)
	}

	if err != nil {
		return nil, errorx.NewDefaultError("查询配置列表失败")
	}

	// Filter by key if provided (in-memory filter)
	var filteredConfigs []*model.MdConfiguration
	if req.Key != "" {
		for _, config := range configs {
			if config.ConfigKey == req.Key {
				filteredConfigs = append(filteredConfigs, config)
			}
		}
		configs = filteredConfigs
		total = int64(len(filteredConfigs))
	}

	// Convert to response types
	var list []types.Configuration
	for _, c := range configs {
		desc := ""
		if c.Description.Valid {
			desc = c.Description.String
		}

		list = append(list, types.Configuration{
			Id:             c.Id,
			Module:         c.Module,
			Key:            c.ConfigKey,
			Value:          c.ConfigValue,
			ValueType:      c.ValueType,
			Description:    desc,
			IsPublic:       int32(c.IsPublic),
			ApprovalStatus: int32(c.ApprovalStatus),
			CreatedBy:      c.CreatedBy,
			CreatedTime:    c.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:    c.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetConfigurationsResp{
		List:  list,
		Total: total,
	}, nil
}
