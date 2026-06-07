package user

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUsersLogic) ListUsers(in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	page := in.Page.GetPage()
	pageSize := in.Page.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	keyword := ""
	if in.Keyword != nil {
		keyword = *in.Keyword
	}

	var status *int64
	if in.Status != nil {
		s := int64(*in.Status)
		status = &s
	}

	users, total, err := l.svcCtx.UserBaseModel.FindPage(l.ctx, keyword, status, page, pageSize)
	if err != nil {
		l.Errorf("list users error: %v", err)
		return nil, err
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &userv1.ListUsersResponse{
		Base:  responsex.NewBaseResp(),
		Users: toProtoUsers(users),
		Page: &commonv1.PageResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
