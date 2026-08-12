package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ListPermissions 查询权限树
//
//	查 sys_permission 全表 → 按 parent_id 构建树形结构
func (l *ListPermissionsLogic) ListPermissions(in *permissionv1.ListPermissionsRequest) (*permissionv1.ListPermissionsResponse, error) {
	// 转换请求中的可选筛选参数（int32 → int64）
	var typeFilter, statusFilter *int64
	if in.Type != nil {
		t := int64(*in.Type)
		typeFilter = &t
	}
	if in.Status != nil {
		s := int64(*in.Status)
		statusFilter = &s
	}

	perms, err := l.svcCtx.PermissionModel.FindWithFilter(l.ctx, typeFilter, statusFilter)
	if err != nil {
		return nil, err
	}

	// 构建 map[id]*Permission
	permMap := make(map[int64]*permissionv1.Permission)
	for _, p := range perms {
		permMap[p.Id] = &permissionv1.Permission{
			Id:           p.Id,
			ParentId:     p.ParentId.Int64,
			Code:         p.Code,
			Name:         p.Name,
			Type:         int32(p.Type),
			Path:         p.Path.String,
			Icon:         p.Icon.String,
			SortOrder:    int32(p.SortOrder),
			Status:       int32(p.Status),
			MinVerfLevel: int32(p.MinVerfLevel),
		}
	}

	// 构建树
	var roots []*permissionv1.Permission
	for _, p := range perms {
		pb := permMap[p.Id]
		if p.ParentId.Valid && p.ParentId.Int64 > 0 {
			if parent, ok := permMap[p.ParentId.Int64]; ok {
				parent.Children = append(parent.Children, pb)
			}
		} else {
			roots = append(roots, pb)
		}
	}

	return &permissionv1.ListPermissionsResponse{
		Base:        responsex.NewBaseResp(),
		Permissions: roots,
	}, nil
}
