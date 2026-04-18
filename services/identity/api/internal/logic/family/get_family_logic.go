package family

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFamilyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFamilyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFamilyLogic {
	return &GetFamilyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFamilyLogic) GetFamily(req *types.GetFamilyReq) (resp *types.GetFamilyResp, err error) {
	family, err := l.svcCtx.AuthFamilyModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewDefaultError("family not found")
		}
		logx.Errorf("failed to get family: %v", err)
		return nil, errorx.NewDefaultError("failed to get family")
	}

	members, err := l.svcCtx.AuthFamilyMemberModel.FindByFamilyId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("failed to get family members: %v", err)
		return nil, errorx.NewDefaultError("failed to get family members")
	}

	memberList := make([]types.FamilyMember, 0, len(members))
	for _, m := range members {
		member := types.FamilyMember{
			Id:           m.Id,
			FamilyId:     m.FamilyId,
			Name:         m.Name,
			Relationship: m.Relationship,
			CreatedTime:  m.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedTime:  m.UpdatedTime.Format("2006-01-02T15:04:05Z07:00"),
		}
		if m.UserId.Valid {
			userId := m.UserId.Int64
			member.UserId = &userId
		}
		if m.Phone.Valid {
			phone := m.Phone.String
			member.Phone = &phone
		}
		if m.IdCardNumber.Valid {
			idCard := m.IdCardNumber.String
			member.IdCardNumber = &idCard
		}
		if m.BirthDate.Valid {
			birthDate := m.BirthDate.Time.Format("2006-01-02")
			member.BirthDate = &birthDate
		}
		if m.Gender.Valid {
			gender := m.Gender.Int64
			member.Gender = &gender
		}
		memberList = append(memberList, member)
	}

	familyName := ""
	if family.FamilyName.Valid {
		familyName = family.FamilyName.String
	}

	return &types.GetFamilyResp{
		Family: types.FamilyWithMembers{
			Id:             family.Id,
			PropertyUnitId: family.PropertyUnitId,
			FamilyHeadId:   family.FamilyHeadId,
			FamilyName:     familyName,
			Status:         family.Status,
			Members:        memberList,
			CreatedTime:    family.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedTime:    family.UpdatedTime.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}
