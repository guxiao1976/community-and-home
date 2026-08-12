package perm

import (
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/types"
)

// toRoleInfo 将 proto Role 转换为 HTTP RoleInfo
func toRoleInfo(r *permissionv1.Role) types.RoleInfo {
	info := types.RoleInfo{
		Id:          r.Id,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		Status:      r.Status,
		SortOrder:   r.SortOrder,
		Permissions: toPermissionInfoList(r.Permissions),
	}
	if r.Timestamps != nil {
		info.CreatedAt = r.Timestamps.CreatedAt
		info.UpdatedAt = r.Timestamps.UpdatedAt
	}
	return info
}

// toPermissionInfo 将 proto Permission 转换为 HTTP PermissionInfo（含子节点递归）
func toPermissionInfo(p *permissionv1.Permission) types.PermissionInfo {
	info := types.PermissionInfo{
		Id:           p.Id,
		ParentId:     p.ParentId,
		Code:         p.Code,
		Name:         p.Name,
		Type:         p.Type,
		Path:         p.Path,
		Icon:         p.Icon,
		SortOrder:    p.SortOrder,
		Status:       p.Status,
		MinVerfLevel: p.MinVerfLevel,
		Children:     toPermissionInfoList(p.Children),
	}
	if p.Timestamps != nil {
		info.CreatedAt = p.Timestamps.CreatedAt
		info.UpdatedAt = p.Timestamps.UpdatedAt
	}
	return info
}

// toPermissionInfoList 批量转换
func toPermissionInfoList(perms []*permissionv1.Permission) []types.PermissionInfo {
	if len(perms) == 0 {
		return nil
	}
	result := make([]types.PermissionInfo, 0, len(perms))
	for _, p := range perms {
		result = append(result, toPermissionInfo(p))
	}
	return result
}
