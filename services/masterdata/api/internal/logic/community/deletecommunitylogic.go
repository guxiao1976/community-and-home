package community

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommunityLogic {
	return &DeleteCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommunityLogic) DeleteCommunity(req *types.DeleteCommunityReq) (resp *types.DeleteCommunityResp, err error) {
	// 1. Find existing
	existing, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("社区不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("社区已删除")
	}

	// 2. Soft delete
	existing.DeleteTime = sql.NullTime{Time: time.Now(), Valid: true}
	existing.UpdatedTime = time.Now()

	if err := l.svcCtx.MdCommunityModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("删除社区失败")
	}

	// 3. Create audit log
	userId := int64(0) // TODO: Get from JWT
	auditLog := &model.MdAuditLog{
		UserId:      userId,
		EntityType:  "md_community",
		EntityId:    existing.Id,
		Action:      "DELETE",
		IpAddress:   "0.0.0.0", // TODO: Get from request
		CreatedTime: time.Now(),
	}
	_, _ = l.svcCtx.MdAuditLogModel.Insert(l.ctx, auditLog)

	return &types.DeleteCommunityResp{Success: true}, nil
}