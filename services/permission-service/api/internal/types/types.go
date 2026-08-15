// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package types

import (
	"encoding/json"
	"strconv"
)

// SEE: [[proto-jstype]] — Int64Array 包装类型，将 []int64 序列化为字符串数组以避免 JavaScript 精度丢失
type Int64Array []int64

func (a Int64Array) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	strs := make([]string, len(a))
	for i, v := range a {
		strs[i] = strconv.FormatInt(v, 10)
	}
	return json.Marshal(strs)
}

func (a *Int64Array) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	*a = make([]int64, len(strs))
	for i, s := range strs {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		(*a)[i] = v
	}
	return nil
}

// ==================== Common ====================

// PageInfo 分页信息
type PageInfo struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"pageSize"`
	Total      int64 `json:"total,string"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
	TotalPages int32 `json:"totalPages"`
}

// ==================== Role ====================

// RoleInfo 角色信息（HTTP 响应）
type RoleInfo struct {
	Id          int64  `json:"id,string"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"isSystem"`
	Status      int32  `json:"status"`
	SortOrder   int32  `json:"sortOrder"`
	// Platforms 允许登录端（pc/mobile）；无 omitempty：空必须序列化为 [] 非 null，供前端 Array.isArray 判定「全部」
	// SEE: [[edit-form-data-integrity]] — 编辑回显完整（8 层链路）
	Platforms   []string         `json:"platforms"`
	Permissions []PermissionInfo `json:"permissions,omitempty"`
	CreatedAt   int64            `json:"createdAt,string"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
	UpdatedAt   int64            `json:"updatedAt,string"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
}

// ListRolesReq 角色列表请求
type ListRolesReq struct {
	Page      int64   `form:"page,default=1"`
	PageSize  int64   `form:"pageSize,default=10"`
	Status    *int32  `form:"status,optional"`
	SortBy    *string `form:"sortBy,optional"`    // 排序字段（白名单校验在 RPC 层）
	SortOrder *string `form:"sortOrder,optional"` // 排序方向 asc/desc，空串由 RPC 层默认 asc
}

// ListRolesResp 角色列表响应
type ListRolesResp struct {
	Roles []RoleInfo `json:"roles"`
	Page  PageInfo   `json:"page"`
}

// CreateRoleReq 创建角色请求
type CreateRoleReq struct {
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Description   string     `json:"description,optional"`
	SortOrder     int32      `json:"sortOrder,optional"`
	PermissionIds Int64Array `json:"permissionIds,optional"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
	Platforms     []string   `json:"platforms,optional"`     // 允许登录端：pc/mobile；空=fail-open
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
	Id            int64      `path:"id"`
	Name          *string    `json:"name,optional"`
	Description   *string    `json:"description,optional"`
	Status        *int32     `json:"status,optional"`
	SortOrder     *int32     `json:"sortOrder,optional"`
	PermissionIds Int64Array `json:"permissionIds,optional"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
	Platforms     []string   `json:"platforms,optional"`     // 允许登录端：pc/mobile；空=fail-open（D3 恒透传，空列表=显式清空）
}

// UpdateRoleResp 更新角色响应
type UpdateRoleResp struct {
	Role RoleInfo `json:"role"`
}

// DeleteRoleReq 删除角色请求（path param）
type DeleteRoleReq struct {
	Id int64 `path:"id"`
}

// ==================== Role Users ====================

type ListRoleUsersReq struct {
	Id       int64 `path:"id"`
	Page     int32 `form:"page,optional,default=1"`
	PageSize int32 `form:"pageSize,optional,default=20"`
}

type RoleUserInfo struct {
	UserId   int64  `json:"userId,string"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Status   int32  `json:"status"` // 用户状态：1-启用 2-禁用（与 user_base.status 一致）
}

type ListRoleUsersResp struct {
	Users      []RoleUserInfo `json:"users"`
	Page       int32          `json:"page"`
	PageSize   int32          `json:"pageSize"`
	Total      int64          `json:"total,string"`
	TotalPages int32          `json:"totalPages"`
}

// ==================== Permission ====================

// PermissionInfo 权限信息（HTTP 响应，支持树形结构）
type PermissionInfo struct {
	Id           int64            `json:"id,string"`
	ParentId     int64            `json:"parentId,string"`
	Code         string           `json:"code"`
	Name         string           `json:"name"`
	Type         int32            `json:"type"`
	Path         string           `json:"path"`
	Icon         string           `json:"icon"`
	SortOrder    int32            `json:"sortOrder"`
	Status       int32            `json:"status"`
	MinVerfLevel int32            `json:"minVerfLevel"` // T1.1 能力层级透出：0=持角色+数据范围即可, 2=需已认证
	Children     []PermissionInfo `json:"children,omitempty"`
	CreatedAt    int64            `json:"createdAt,string"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
	UpdatedAt    int64            `json:"updatedAt,string"` // SEE: [[proto-jstype]] — 避免 JavaScript Number 精度丢失
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

// ==================== Role-Permission Assignment ====================

// GetRolePermissionsResp 获取角色权限 ID 列表响应
type GetRolePermissionsResp struct {
	PermissionIds Int64Array `json:"permissionIds"`
}

// AssignRolePermissionsReq 分配角色权限请求
type AssignRolePermissionsReq struct {
	PermissionIds Int64Array `json:"permissionIds"`
}

// ==================== User-Role Assignment ====================

// AssignUserRoleReq 为用户分配角色（管理员操作，含数据范围）
type AssignUserRoleReq struct {
	UserId    int64  `json:"userId,string"`
	RoleId    int64  `json:"roleId,string"`
	ScopeType string `json:"scopeType"`
	ScopeId   int64  `json:"scopeId,string"`
}

// RevokeUserRoleReq 撤销用户角色（管理员操作）
type RevokeUserRoleReq struct {
	UserId    int64  `json:"userId,string"`
	RoleId    int64  `json:"roleId,string"`
	ScopeType string `json:"scopeType,optional"`
	ScopeId   int64  `json:"scopeId,string,optional"`
}

// UserRoleInfo 用户角色关联信息（HTTP 响应）
type UserRoleInfo struct {
	Role       RoleInfo `json:"role"`
	ScopeType  string   `json:"scopeType"`
	ScopeId    int64    `json:"scopeId,string"`
	Status     int32    `json:"status"`            // 个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期
	VerifiedAt int64    `json:"verifiedAt,string"` // 认证通过时间
	ExpiresAt  int64    `json:"expiresAt,string"`  // 到期时间, 0=永久
}

// GetUserRolesReq 查询用户角色请求（path param）
type GetUserRolesReq struct {
	UserId int64 `path:"userId"`
}

// GetUserRolesResp 查询用户角色响应
type GetUserRolesResp struct {
	Roles []UserRoleInfo `json:"roles"`
}

// GetUserPermissionsReq 查询用户权限请求（path param）
type GetUserPermissionsReq struct {
	UserId int64 `path:"userId"`
}

// GetUserPermissionsResp 查询用户权限响应
type GetUserPermissionsResp struct {
	PermissionCodes []string `json:"permissionCodes"`
}

// ==================== Auto-Discover ====================

// AutoDiscoveredPerm 自动发现的单条权限
type AutoDiscoveredPerm struct {
	Id       int64  `json:"id,string"`
	ParentId int64  `json:"parentId,string"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Path     string `json:"path"`
}

// AutoDiscoverPermissionsResp 自动发现权限响应
type AutoDiscoverPermissionsResp struct {
	Added   []AutoDiscoveredPerm `json:"added"`
	Total   int                  `json:"total"`
	Message string               `json:"message"`
}
