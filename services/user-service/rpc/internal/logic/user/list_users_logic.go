package user

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
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

	maxPageSize := int32(100)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.list.max_page_size"); err == nil {
			maxPageSize = int32(v)
		}
	}
	defaultPageSize := int32(10)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.list.default_page_size"); err == nil {
			defaultPageSize = int32(v)
		}
	}

	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	keyword := ""
	if in.Keyword != nil {
		keyword = *in.Keyword
	}
	encryptedPhone := ""
	if in.Phone != nil && *in.Phone != "" {
		enc, err := crypto.AESEncrypt(*in.Phone)
		if err != nil {
			l.Errorf("AES encrypt phone failed: %v", err)
			return nil, err
		}
		encryptedPhone = enc
	}

	var status *int64
	if in.Status != nil {
		s := int64(*in.Status)
		status = &s
	}

	users, total, err := l.svcCtx.UserBaseModel.FindPage(l.ctx, keyword, encryptedPhone, status, page, pageSize)
	if err != nil {
		l.Errorf("list users error: %v", err)
		return nil, err
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	pbUsers := toProtoUsers(users)
	l.fillRoleNames(l.ctx, pbUsers)

	return &userv1.ListUsersResponse{
		Base:  responsex.NewBaseResp(),
		Users: pbUsers,
		Page: &commonv1.PageResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// fillRoleNames 逐用户从 permission-service 聚合已分配角色名称（列表展示用）。
// 单用户角色获取失败不影响整体列表（失败时该用户 role_names 为空）。
func (l *ListUsersLogic) fillRoleNames(ctx context.Context, users []*userv1.User) {
	if len(users) == 0 || l.svcCtx.PermissionClient == nil {
		return
	}
	for _, u := range users {
		if u == nil || u.Id == 0 {
			continue
		}
		resp, err := l.svcCtx.PermissionClient.GetUserRoles(ctx, &permissionv1.GetUserRolesRequest{UserId: u.Id})
		if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			continue
		}
		names := make([]string, 0, len(resp.Roles))
		seen := make(map[string]bool)
		for _, ur := range resp.Roles {
			if ur != nil && ur.Role != nil && ur.Role.Name != "" && !seen[ur.Role.Name] {
				seen[ur.Role.Name] = true
				names = append(names, ur.Role.Name)
			}
		}
		u.RoleNames = names
	}
}
