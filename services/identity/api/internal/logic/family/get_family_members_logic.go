package family

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFamilyMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFamilyMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFamilyMembersLogic {
	return &GetFamilyMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFamilyMembersLogic) GetFamilyMembers(req *types.GetFamilyMembersReq) (resp *types.GetFamilyMembersResp, err error) {
	members, err := l.svcCtx.AuthFamilyMemberModel.FindByFamilyId(l.ctx, req.FamilyId)
	if err != nil {
		logx.Errorf("failed to get family members: %v", err)
		return nil, errorx.NewDefaultError("failed to get family members")
	}

	list := make([]types.FamilyMember, 0, len(members))
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
		list = append(list, member)
	}

	return &types.GetFamilyMembersResp{
		List: list,
	}, nil
}
