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

type AddFamilyMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFamilyMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFamilyMemberLogic {
	return &AddFamilyMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFamilyMemberLogic) AddFamilyMember(req *types.AddFamilyMemberReq) (resp *types.AddFamilyMemberResp, err error) {
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	family, err := l.svcCtx.AuthFamilyModel.FindOne(l.ctx, req.FamilyId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("family not found")
		}
		logx.Errorf("failed to get family: %v", err)
		return nil, errorx.NewDefaultError("failed to get family")
	}

	if family.FamilyHeadId != userId {
		return nil, errorx.NewDefaultError("only family head can add members")
	}

	now := time.Now()
	member := &model.AuthFamilyMember{
		FamilyId:     req.FamilyId,
		Name:         req.Name,
		Relationship: req.Relationship,
		CreatedTime:  now,
		UpdatedTime:  now,
	}

	if req.UserId != nil {
		member.UserId = sql.NullInt64{Int64: *req.UserId, Valid: true}
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

	result, err := l.svcCtx.AuthFamilyMemberModel.Insert(l.ctx, member)
	if err != nil {
		logx.Errorf("failed to add family member: %v", err)
		return nil, errorx.NewDefaultError("failed to add family member")
	}

	memberId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get member id: %v", err)
		return nil, errorx.NewDefaultError("failed to add family member")
	}

	return &types.AddFamilyMemberResp{
		Id: memberId,
	}, nil
}
