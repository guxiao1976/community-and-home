package amap_sync

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSyncProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncProgressLogic {
	return &GetSyncProgressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSyncProgressLogic) GetSyncProgress(req *types.GetSyncProgressReq) (resp *types.GetSyncProgressResp, err error) {
	if req.TaskId == "" {
		return nil, errorx.NewInvalidParamError("task_id is required")
	}

	progress := l.svcCtx.SyncEngine.GetProgress(req.TaskId)
	if progress == nil {
		return nil, errorx.NewDefaultError("sync task not found")
	}

	return &types.GetSyncProgressResp{
		TaskId:            progress.TaskId,
		Status:            string(progress.Status),
		TotalCounties:     progress.TotalCounties,
		CurrentCounty:     progress.CurrentCountyIdx,
		CurrentCountyName: progress.CurrentCountyName,
		TotalKeywords:     progress.TotalKeywords,
		CurrentKeyword:    progress.CurrentKeywordIdx,
		CurrentKeywordStr: progress.CurrentKeyword,
		TotalPages:        progress.TotalPages,
		CurrentPage:       progress.CurrentPage,
		TotalFound:        progress.TotalFound,
		TotalSynced:       progress.TotalSynced,
		TotalSkipped:      progress.TotalSkipped,
		TotalFailed:       progress.TotalFailed,
		ErrorMessage:      progress.ErrorMessage,
	}, nil
}
