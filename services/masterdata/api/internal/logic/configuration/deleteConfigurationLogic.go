// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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

type DeleteConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Soft delete configuration
func NewDeleteConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConfigurationLogic {
	return &DeleteConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteConfigurationLogic) DeleteConfiguration(req *types.DeleteConfigurationReq) (resp *types.DeleteConfigurationResp, err error) {
	config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("配置不存在")
		}
		return nil, errorx.NewDefaultError("查询配置失败")
	}
	if config.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("配置已删除")
	}

	config.SubmissionStatus = 4
	config.SubmissionType = sql.NullInt64{Int64: 3, Valid: true}
	if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, config); err != nil {
		return nil, errorx.NewDefaultError("删除配置失败")
	}

	return &types.DeleteConfigurationResp{Success: true}, nil
}
