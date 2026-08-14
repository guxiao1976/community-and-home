package perm

import (
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/types"
)

// toRoleInfo 将 proto Role 转换为 HTTP RoleInfo
//
//	REQ-UPDATE-2：nil 防御 — r==nil 返回零值 types.RoleInfo{} 不 panic
//	REQ-PLAT-6：platforms 透传；nil/空归一为 []string{}（空数组非 null，供前端 Array.isArray 判定「全部」）
func toRoleInfo(r *permissionv1.Role) types.RoleInfo {
	if r == nil {
		return types.RoleInfo{}
	}
	info := types.RoleInfo{
		Id:          r.Id,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		Status:      r.Status,
		SortOrder:   r.SortOrder,
		Platforms:   normalizePlatforms(r.Platforms),
		Permissions: toPermissionInfoList(r.Permissions),
	}
	if r.Timestamps != nil {
		info.CreatedAt = r.Timestamps.CreatedAt
		info.UpdatedAt = r.Timestamps.UpdatedAt
	}
	return info
}

// normalizePlatforms 将 proto platforms 归一为空数组（nil/空 → []string{}，保持非 nil）
// SEE: [[edit-form-data-integrity]] — 空平台必须序列化为 [] 而非 null，前端编辑回显完整
func normalizePlatforms(ps []string) []string {
	if len(ps) == 0 {
		return []string{}
	}
	return ps
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
