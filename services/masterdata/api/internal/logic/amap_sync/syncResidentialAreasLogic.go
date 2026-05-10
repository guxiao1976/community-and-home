package amap_sync

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncResidentialAreasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncResidentialAreasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncResidentialAreasLogic {
	return &SyncResidentialAreasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncResidentialAreasLogic) SyncResidentialAreas(req *types.SyncResidentialAreasReq) (resp *types.SyncResidentialAreasResp, err error) {
	if req.DivisionId == 0 {
		return nil, errorx.NewInvalidParamError("division_id is required")
	}

	taskId := l.svcCtx.SyncEngine.StartSync(l.ctx, req.DivisionId)
	return &types.SyncResidentialAreasResp{TaskId: taskId}, nil
}
