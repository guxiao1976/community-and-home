// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package types

// ==================== Common ====================

// PageInfo 分页信息
type PageInfo struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int32 `json:"totalPages"`
}

// ==================== Role ====================

// RoleInfo 角色信息（HTTP 响应）
type RoleInfo struct {
	Id          int64            `json:"id,string"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	IsSystem    bool             `json:"isSystem"`
	Status      int32            `json:"status"`
	SortOrder   int32            `json:"sortOrder"`
	Permissions []PermissionInfo `json:"permissions,omitempty"`
	CreatedAt   int64            `json:"createdAt"`
	UpdatedAt   int64            `json:"updatedAt"`
}

// ListRolesReq 角色列表请求
type ListRolesReq struct {
	Page     int64  `form:"page,default=1"`
	PageSize int64  `form:"pageSize,default=10"`
	Status   *int32 `form:"status,optional"`
}

// ListRolesResp 角色列表响应
type ListRolesResp struct {
	Roles []RoleInfo `json:"roles"`
	Page  PageInfo   `json:"page"`
}

// CreateRoleReq 创建角色请求
type CreateRoleReq struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Description   string  `json:"description,optional"`
	SortOrder     int32   `json:"sortOrder,optional"`
	PermissionIds []int64 `json:"permissionIds,optional"`
}

// CreateRoleResp 创建角色响应
type CreateRoleResp struct {
	Role RoleInfo `json:"role"`
}

// GetRoleReq 获取角色请求（path param）
type GetRoleReq struct {
	Id int64 `path:"id"`
}

// GetRoleResp 获取角色响应
type GetRoleResp struct {
	Role RoleInfo `json:"role"`
}

// UpdateRoleReq 更新角色请求（path param + body）
type UpdateRoleReq struct {
	Id            int64   `path:"id"`
	Name          *string `json:"name,optional"`
	Description   *string `json:"description,optional"`
	Status        *int32  `json:"status,optional"`
	SortOrder     *int32  `json:"sortOrder,optional"`
	PermissionIds []int64 `json:"permissionIds,optional"`
}

// UpdateRoleResp 更新角色响应
type UpdateRoleResp struct {
	Role RoleInfo `json:"role"`
}

// DeleteRoleReq 删除角色请求（path param）
type DeleteRoleReq struct {
	Id int64 `path:"id"`
}

// ==================== Permission ====================

// PermissionInfo 权限信息（HTTP 响应，支持树形结构）
type PermissionInfo struct {
	Id        int64            `json:"id,string"`
	ParentId  int64            `json:"parentId,string"`
	Code      string           `json:"code"`
	Name      string           `json:"name"`
	Type      int32            `json:"type"`
	Path      string           `json:"path"`
	Icon      string           `json:"icon"`
	SortOrder int32            `json:"sortOrder"`
	Status    int32            `json:"status"`
	Children  []PermissionInfo `json:"children,omitempty"`
	CreatedAt int64            `json:"createdAt"`
	UpdatedAt int64            `json:"updatedAt"`
}

// ListPermissionsReq 权限列表请求
type ListPermissionsReq struct {
	Type   *int32 `form:"type,optional"`
	Status *int32 `form:"status,optional"`
}

// ListPermissionsResp 权限列表响应（树形结构）
type ListPermissionsResp struct {
	Permissions []PermissionInfo `json:"permissions"`
}
