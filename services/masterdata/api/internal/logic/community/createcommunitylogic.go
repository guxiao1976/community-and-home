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

type CreateCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommunityLogic {
	return &CreateCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCommunityLogic) CreateCommunity(req *types.CreateCommunityReq) (resp *types.CreateCommunityResp, err error) {
	// 1. Validate division exists and is level 5 (community level)
	division, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.DivisionId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewInvalidParamError("行政区域不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区域失败")
	}
	if division.Level != 5 {
		return nil, errorx.NewInvalidParamError("社区必须关联到第5级(社区级)行政区域")
	}

	// 2. Create community model
	data := &model.MdCommunity{
		DivisionId:       req.DivisionId,
		Name:             req.Name,
		Address:          req.Address,
		CommunityType:    int64(req.CommunityType),
		SubmissionStatus: 0, // Draft
		SubmitterId:      0, // TODO: Get from JWT
		CreatedTime:      time.Now(),
		UpdatedTime:      time.Now(),
	}
	if req.Area != 0 {
		data.Area = sql.NullFloat64{Float64: req.Area, Valid: true}
	}
	if req.Population != 0 {
		data.Population = sql.NullInt64{Int64: int64(req.Population), Valid: true}
	}

	res, err := l.svcCtx.MdCommunityModel.Insert(l.ctx, data)
	if err != nil {
		return nil, errorx.NewDefaultError("创建社区失败")
	}

	id, _ := res.LastInsertId()
	return &types.CreateCommunityResp{Id: id}, nil
}