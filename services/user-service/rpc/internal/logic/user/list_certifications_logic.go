package user

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type ListCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCertificationsLogic {
	return &ListCertificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCertificationsLogic) ListCertifications(in *userv1.ListCertificationsRequest) (*userv1.ListCertificationsResponse, error) {
	page := in.Page.GetPage()
	pageSize := in.Page.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var status *int64
	if in.Status != nil {
		s := int64(*in.Status)
		status = &s
	}

	var userId *int64
	if in.UserId != nil {
		uid := *in.UserId
		userId = &uid
	}

	certs, total, err := l.svcCtx.UserCertificationModel.FindPage(l.ctx, status, userId, page, pageSize)
	if err != nil {
		l.Errorf("list certifications error: %v", err)
		return nil, err
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &userv1.ListCertificationsResponse{
		Base:           responsex.NewBaseResp(),
		Certifications: toProtoCertifications(certs),
		Page: &commonv1.PageResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
