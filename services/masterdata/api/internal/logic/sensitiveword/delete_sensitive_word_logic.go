package sensitiveword

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSensitiveWordLogic {
	return &DeleteSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSensitiveWordLogic) DeleteSensitiveWord(req *types.DeleteSensitiveWordReq) (resp *types.DeleteSensitiveWordResp, err error) {
	_, err = l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("sensitive word not found")
	}

	err = l.svcCtx.MdSensitiveWordModel.Delete(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to delete sensitive word")
	}

	return &types.DeleteSensitiveWordResp{
		Success: true,
	}, nil
}
