package configuration

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConfigLogic {
	return &DeleteConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteConfigLogic) DeleteConfig(req *types.DeleteConfigReq) (resp *types.DeleteConfigResp, err error) {
	_, err = l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("configuration not found")
	}

	err = l.svcCtx.MdConfigurationModel.Delete(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to delete configuration")
	}

	return &types.DeleteConfigResp{
		Success: true,
	}, nil
}
