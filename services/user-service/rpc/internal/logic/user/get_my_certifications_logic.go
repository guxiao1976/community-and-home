package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetMyCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCertificationsLogic {
	return &GetMyCertificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMyCertificationsLogic) GetMyCertifications(in *userv1.GetMyCertificationsRequest) (*userv1.GetMyCertificationsResponse, error) {
	certs, err := l.svcCtx.UserCertificationModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("find certifications error: %v", err)
		return nil, err
	}

	return &userv1.GetMyCertificationsResponse{
		Base:           responsex.NewBaseResp(),
		Certifications: toProtoCertifications(certs),
	}, nil
}
