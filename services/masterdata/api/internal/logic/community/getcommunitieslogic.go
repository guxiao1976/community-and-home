package community

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommunitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCommunitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommunitiesLogic {
	return &GetCommunitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCommunitiesLogic) GetCommunities(req *types.GetCommunitiesReq) (resp *types.GetCommunitiesResp, err error) {
	var communities []*model.MdCommunity

	if req.DivisionId != nil {
		communities, err = l.svcCtx.MdCommunityModel.FindByDivisionId(l.ctx, *req.DivisionId)
	} else {
		communities, err = l.svcCtx.MdCommunityModel.FindAll(l.ctx)
	}
	if err != nil {
		return nil, errorx.NewDefaultError("查询社区列表失败")
	}

	// Filter by submission status if requested
	var list []types.Community
	for _, c := range communities {
		if c.DeleteTime.Valid {
			continue
		}
		if req.SubmissionStatus != nil && c.SubmissionStatus != int64(*req.SubmissionStatus) {
			continue
		}
		list = append(list, toCommunityType(c))
	}

	return &types.GetCommunitiesResp{List: list, Total: int64(len(list))}, nil
}

func toCommunityType(c *model.MdCommunity) types.Community {
	result := types.Community{
		Id:               c.Id,
		DivisionId:       c.DivisionId,
		Name:             c.Name,
		Address:          c.Address,
		CommunityType:    int32(c.CommunityType),
		SubmissionStatus: int32(c.SubmissionStatus),
		SubmitterId:      c.SubmitterId,
		CreatedTime:      c.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime:      c.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
	if c.Area.Valid {
		result.Area = c.Area.Float64
	}
	if c.Population.Valid {
		result.Population = int32(c.Population.Int64)
	}
	if c.SubmitTime.Valid {
		result.SubmitTime = c.SubmitTime.Time.Format("2006-01-02 15:04:05")
	}
	if c.ReviewerId.Valid {
		rid := c.ReviewerId.Int64
		result.ReviewerId = &rid
	}
	if c.ReviewTime.Valid {
		result.ReviewTime = c.ReviewTime.Time.Format("2006-01-02 15:04:05")
	}
	if c.ReviewNotes.Valid {
		result.ReviewNotes = c.ReviewNotes.String
	}
	return result
}