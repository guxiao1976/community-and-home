package family

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFamilyMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFamilyMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFamilyMemberLogic {
	return &DeleteFamilyMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFamilyMemberLogic) DeleteFamilyMember(req *types.DeleteFamilyMemberReq) (resp *types.DeleteFamilyMemberResp, err error) {
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	member, err := l.svcCtx.AuthFamilyMemberModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("member not found")
		}
		logx.Errorf("failed to get member: %v", err)
		return nil, errorx.NewDefaultError("failed to get member")
	}

	family, err := l.svcCtx.AuthFamilyModel.FindOne(l.ctx, member.FamilyId)
	if err != nil {
		logx.Errorf("failed to get family: %v", err)
		return nil, errorx.NewDefaultError("failed to get family")
	}

	if family.FamilyHeadId != userId {
		return nil, errorx.NewDefaultError("only family head can delete members")
	}

	err = l.svcCtx.AuthFamilyMemberModel.Delete(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to delete member: %v", err)
		return nil, errorx.NewDefaultError("failed to delete member")
	}

	return &types.DeleteFamilyMemberResp{
		Success: true,
	}, nil
}
