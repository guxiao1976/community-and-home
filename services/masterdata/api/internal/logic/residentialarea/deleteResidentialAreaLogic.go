package residentialarea

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteResidentialAreaLogic {
	return &DeleteResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteResidentialAreaLogic) DeleteResidentialArea(req *types.DeleteResidentialAreaReq) (resp *types.DeleteResidentialAreaResp, err error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	if area.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("住宅小区已删除")
	}
	if area.SubmissionStatus == 4 {
		return nil, errorx.NewDefaultError("住宅小区已在待删除状态")
	}

	area.SubmissionStatus = 4 // 标记为待删除，需提交审批
	if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
		return nil, errorx.NewDefaultError("删除住宅小区失败: " + err.Error())
	}

	return &types.DeleteResidentialAreaResp{Success: true}, nil
}
