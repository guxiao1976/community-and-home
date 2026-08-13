package permission

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ListRoles 分页查询角色列表
func (l *ListRolesLogic) ListRoles(in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
	page := int64(1)
	pageSize := int64(10)
	if in.Page != nil {
		if in.Page.Page > 0 {
			page = int64(in.Page.Page)
		}
		if in.Page.PageSize > 0 {
			pageSize = int64(in.Page.PageSize)
		}
	}
	var status *int64
	if in.Status != nil {
		v := int64(*in.Status)
		status = &v
	}

	roles, total, err := l.svcCtx.RoleModel.FindList(l.ctx, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	var pbRoles []*permissionv1.Role
	for _, r := range roles {
		pbRoles = append(pbRoles, &permissionv1.Role{
			Id:          r.Id,
			Code:        r.RoleCode,
			Name:        r.RoleName,
			Description: r.Description.String,
			IsSystem:    r.IsSystem == 1,
			Status:      int32(r.Status),
			SortOrder:   int32(r.SortOrder),
			Platforms:   splitPlatforms(r.Platforms),
		})
	}

	return &permissionv1.ListRolesResponse{
		Base:  responsex.NewBaseResp(),
		Roles: pbRoles,
		Page: &commonv1.PageResponse{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			Total:      total,
			TotalPages: int32((total + pageSize - 1) / pageSize),
		},
	}, nil
}
