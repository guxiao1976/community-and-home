package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetResidencesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetResidencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidencesLogic {
	return &GetResidencesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetResidencesLogic) GetResidences(in *userv1.GetResidencesRequest) (*userv1.GetResidencesResponse, error) {
	residences, err := l.svcCtx.UserResidenceModel.FindByMembershipId(l.ctx, in.MembershipId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("find residences error: %v", err)
		return nil, err
	}

	result := make([]*userv1.Residence, 0, len(residences))
	for _, r := range residences {
		result = append(result, toProtoResidence(r))
	}

	return &userv1.GetResidencesResponse{
		Base:       responsex.NewBaseResp(),
		Residences: result,
	}, nil
}
