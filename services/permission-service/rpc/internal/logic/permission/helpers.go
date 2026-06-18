package permission

import (
	"context"
	"database/sql"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
)

// sqlNullString converts a string to sql.NullString (empty string → NULL)
func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// toRolePb 将 model.SysRole 转换为 proto Role（不含权限）
func toRolePb(r *model.SysRole) *permissionv1.Role {
	return &permissionv1.Role{
		Id:          r.Id,
		Code:        r.RoleCode,
		Name:        r.RoleName,
		Description: r.Description.String,
		IsSystem:    r.IsSystem == 1,
		Status:      int32(r.Status),
		SortOrder:   int32(r.SortOrder),
		Timestamps: &commonv1.Timestamps{
			CreatedAt: r.CreatedTime.Unix(),
			UpdatedAt: r.UpdatedTime.Unix(),
		},
	}
}

// roleToPbWithPermissions 将 model.SysRole + permissions 转换为 proto Role（含权限列表）
func roleToPbWithPermissions(ctx context.Context, r *model.SysRole, permIds []int64,
	permModel model.SysPermissionModel) *permissionv1.Role {
	pb := toRolePb(r)

	if len(permIds) == 0 {
		return pb
	}

	perms, err := permModel.FindByIds(ctx, permIds)
	if err != nil || len(perms) == 0 {
		return pb
	}

	var pbPerms []*permissionv1.Permission
	for _, p := range perms {
		pbPerms = append(pbPerms, &permissionv1.Permission{
			Id:        p.Id,
			ParentId:  p.ParentId.Int64,
			Code:      p.Code,
			Name:      p.Name,
			Type:      int32(p.Type),
			Path:      p.Path.String,
			Icon:      p.Icon.String,
			SortOrder: int32(p.SortOrder),
			Status:    int32(p.Status),
			Timestamps: &commonv1.Timestamps{
				CreatedAt: p.CreatedTime.Unix(),
				UpdatedAt: p.UpdatedTime.Unix(),
			},
		})
	}
	pb.Permissions = pbPerms
	return pb
}
