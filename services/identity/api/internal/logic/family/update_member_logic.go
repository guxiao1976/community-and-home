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

type UpdateFamilyMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFamilyMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFamilyMemberLogic {
	return &UpdateFamilyMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFamilyMemberLogic) UpdateFamilyMember(req *types.UpdateFamilyMemberReq) (resp *types.UpdateFamilyMemberResp, err error) {
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
		return nil, errorx.NewDefaultError("only family head can update members")
	}

	if req.Name != nil {
		member.Name = *req.Name
	}
	if req.Relationship != nil {
		member.Relationship = *req.Relationship
	}
	if req.Phone != nil {
		member.Phone = sql.NullString{String: *req.Phone, Valid: true}
	}
	if req.IdCardNumber != nil {
		member.IdCardNumber = sql.NullString{String: *req.IdCardNumber, Valid: true}
	}
	if req.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return nil, errorx.NewDefaultError("invalid birth date format")
		}
		member.BirthDate = sql.NullTime{Time: birthDate, Valid: true}
	}
	if req.Gender != nil {
		member.Gender = sql.NullInt64{Int64: *req.Gender, Valid: true}
	}
	member.UpdatedTime = time.Now()

	err = l.svcCtx.AuthFamilyMemberModel.Update(l.ctx, member)
	if err != nil {
		logx.Errorf("failed to update member: %v", err)
		return nil, errorx.NewDefaultError("failed to update member")
	}

	return &types.UpdateFamilyMemberResp{
		Success: true,
	}, nil
}
