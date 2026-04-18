package family

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFamilyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFamilyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFamilyLogic {
	return &UpdateFamilyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFamilyLogic) UpdateFamily(req *types.UpdateFamilyReq) (resp *types.UpdateFamilyResp, err error) {
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	family, err := l.svcCtx.AuthFamilyModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("family not found")
		}
		logx.Errorf("failed to get family: %v", err)
		return nil, errorx.NewDefaultError("failed to get family")
	}

	if family.FamilyHeadId != userId {
		return nil, errorx.NewDefaultError("only family head can update family")
	}

	if req.FamilyName != "" {
		family.FamilyName = sql.NullString{String: req.FamilyName, Valid: true}
	}
	if req.Status != nil {
		family.Status = *req.Status
	}
	family.UpdatedTime = time.Now()

	err = l.svcCtx.AuthFamilyModel.Update(l.ctx, family)
	if err != nil {
		logx.Errorf("failed to update family: %v", err)
		return nil, errorx.NewDefaultError("failed to update family")
	}

	return &types.UpdateFamilyResp{
		Success: true,
	}, nil
}
