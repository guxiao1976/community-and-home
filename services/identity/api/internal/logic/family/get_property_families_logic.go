package family

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPropertyFamiliesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPropertyFamiliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPropertyFamiliesLogic {
	return &GetPropertyFamiliesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPropertyFamiliesLogic) GetPropertyFamilies(req *types.GetPropertyFamiliesReq) (resp *types.GetPropertyFamiliesResp, err error) {
	families, total, err := l.svcCtx.AuthFamilyModel.FindByPropertyUnitId(l.ctx, req.PropertyUnitId, req.Page, req.PageSize, req.Status)
	if err != nil {
		logx.Errorf("failed to get families: %v", err)
		return nil, errorx.NewDefaultError("failed to get families")
	}

	list := make([]types.Family, 0, len(families))
	for _, f := range families {
		familyName := ""
		if f.FamilyName.Valid {
			familyName = f.FamilyName.String
		}

		memberCount, err := l.svcCtx.AuthFamilyMemberModel.CountByFamilyId(l.ctx, f.Id)
		if err != nil {
			logx.Errorf("failed to count family members: %v", err)
			memberCount = 0
		}

		list = append(list, types.Family{
			Id:             f.Id,
			PropertyUnitId: f.PropertyUnitId,
			FamilyHeadId:   f.FamilyHeadId,
			FamilyName:     familyName,
			Status:         f.Status,
			MemberCount:    memberCount,
			CreatedTime:    f.CreatedTime.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedTime:    f.UpdatedTime.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &types.GetPropertyFamiliesResp{
		List:  list,
		Total: total,
	}, nil
}
