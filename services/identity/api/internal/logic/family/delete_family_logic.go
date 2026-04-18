package family

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFamilyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFamilyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFamilyLogic {
	return &DeleteFamilyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFamilyLogic) DeleteFamily(req *types.DeleteFamilyReq) (resp *types.DeleteFamilyResp, err error) {
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
		return nil, errorx.NewDefaultError("only family head can delete family")
	}

	members, err := l.svcCtx.AuthFamilyMemberModel.FindByFamilyId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to get family members: %v", err)
		return nil, errorx.NewDefaultError("failed to delete family")
	}

	for _, member := range members {
		err = l.svcCtx.AuthFamilyMemberModel.Delete(l.ctx, member.Id)
		if err != nil {
			logx.Errorf("failed to delete family member: %v", err)
			return nil, errorx.NewDefaultError("failed to delete family")
		}
	}

	err = l.svcCtx.AuthFamilyModel.Delete(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to delete family: %v", err)
		return nil, errorx.NewDefaultError("failed to delete family")
	}

	return &types.DeleteFamilyResp{
		Success: true,
	}, nil
}
