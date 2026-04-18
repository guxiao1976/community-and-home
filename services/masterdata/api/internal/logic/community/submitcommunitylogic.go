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

type SubmitCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCommunityLogic {
	return &SubmitCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitCommunityLogic) SubmitCommunity(req *types.SubmitCommunityReq) (resp *types.SubmitCommunityResp, err error) {
	// 1. Find existing
	existing, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("社区不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("社区已删除")
	}

	// 2. Validate: current status must be Draft or Rejected
	if existing.SubmissionStatus != 0 && existing.SubmissionStatus != 3 {
		return nil, errorx.NewInvalidParamError("只有草稿或已退回状态的社区可以提交审核")
	}

	// 3. Update status to Submitted (1)
	existing.SubmissionStatus = 1
	existing.SubmitTime = sql.NullTime{Time: time.Now(), Valid: true}
	existing.UpdatedTime = time.Now()

	if err := l.svcCtx.MdCommunityModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("提交审核失败")
	}

	return &types.SubmitCommunityResp{Success: true}, nil
}