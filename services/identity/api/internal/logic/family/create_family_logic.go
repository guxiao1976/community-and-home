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

type CreateFamilyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFamilyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFamilyLogic {
	return &CreateFamilyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFamilyLogic) CreateFamily(req *types.CreateFamilyReq) (resp *types.CreateFamilyResp, err error) {
	userId, ok := l.ctx.Value("userId").(int64)
	if !ok {
		return nil, errorx.NewDefaultError("unauthorized")
	}

	now := time.Now()
	family := &model.AuthFamily{
		PropertyUnitId: req.PropertyUnitId,
		FamilyHeadId:   userId,
		FamilyName:     sql.NullString{String: req.FamilyName, Valid: req.FamilyName != ""},
		Status:         1,
		CreatedTime:    now,
		UpdatedTime:    now,
	}

	result, err := l.svcCtx.AuthFamilyModel.Insert(l.ctx, family)
	if err != nil {
		logx.Errorf("failed to create family: %v", err)
		return nil, errorx.NewDefaultError("failed to create family")
	}

	familyId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("failed to get family id: %v", err)
		return nil, errorx.NewDefaultError("failed to create family")
	}

	return &types.CreateFamilyResp{
		Id: familyId,
	}, nil
}
