// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package residentialarea

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSubmitResidentialAreasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch submit residential areas
func NewBatchSubmitResidentialAreasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSubmitResidentialAreasLogic {
	return &BatchSubmitResidentialAreasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchSubmitResidentialAreasLogic) BatchSubmitResidentialAreas(req *types.BatchSubmitReq) (resp *types.BatchSubmitResp, err error) {
	if len(req.Ids) == 0 {
		return nil, errorx.NewDefaultError("请选择要提交的数据")
	}

	successCount := 0
	for _, id := range req.Ids {
		area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, id)
		if err != nil || area.DeleteTime.Valid {
			continue
		}
		if area.SubmissionStatus != 0 && area.SubmissionStatus != 3 {
			continue
		}
		area.SubmissionStatus = 1
		if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
			continue
		}
		successCount++
	}

	return &types.BatchSubmitResp{Success: successCount > 0}, nil
}
