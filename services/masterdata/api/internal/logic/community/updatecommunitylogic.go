package community

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommunityLogic {
	return &UpdateCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCommunityLogic) UpdateCommunity(req *types.UpdateCommunityReq) (resp *types.UpdateCommunityResp, err error) {
	// 1. Find existing
	existing, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("社区不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("社区已删除")
	}

	// 2. Only allow updates if Draft or Rejected
	if existing.SubmissionStatus != 0 && existing.SubmissionStatus != 3 {
		return nil, errorx.NewInvalidParamError("只有草稿或已退回状态的社区可以编辑")
	}

	// 3. Apply updates
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.Area != 0 {
		existing.Area = sql.NullFloat64{Float64: req.Area, Valid: true}
	}
	if req.Population != 0 {
		existing.Population = sql.NullInt64{Int64: int64(req.Population), Valid: true}
	}
	if req.CommunityType != 0 {
		existing.CommunityType = int64(req.CommunityType)
	}

	existing.UpdatedTime = time.Now()

	// 4. Save
	if err := l.svcCtx.MdCommunityModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("更新社区失败")
	}

	return &types.UpdateCommunityResp{Success: true}, nil
}