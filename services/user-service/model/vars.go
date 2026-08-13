package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ErrNotFound 记录不存在
var ErrNotFound = sqlx.ErrNotFound

// ErrMaxCommunities 最多加入 3 个小区
var ErrMaxCommunities = errors.New("最多加入 3 个小区")

// ErrAlreadyJoined 已加入该小区
var ErrAlreadyJoined = errors.New("已加入该小区")

// ErrAddressAlreadyTaken 同小区同地址已有人加入
var ErrAddressAlreadyTaken = errors.New("该地址已有人加入")

// ErrRoleAlreadyExists 角色已存在
var ErrRoleAlreadyExists = errors.New("该角色已存在")

// ==================== 用户状态 ====================

const (
	UserStatusActive   = 1
	UserStatusDisabled = 2
)

// ==================== 小区成员绑定状态 ====================

const (
	MembershipBindStatusActive = 1
	MembershipBindStatusLeft   = 0
)

// ==================== 角色认证状态 ====================

const (
	RoleVerfStatusUnverified = 0 // 未认证
	RoleVerfStatusPending    = 1 // 待审核
	RoleVerfStatusApproved   = 2 // 已通过
	RoleVerfStatusRejected   = 3 // 已驳回
	RoleVerfStatusExpired    = 4 // 已过期
)

// ==================== 认证审核状态 ====================

const (
	CertStatusPending  = 1 // 待审核
	CertStatusApproved = 2 // 已通过
	CertStatusRejected = 3 // 已驳回
)

// ==================== 角色编码 ====================

const (
	RoleCodeOwner          = "owner"
	RoleCodeTenant         = "tenant"
	RoleCodeGridWorker     = "grid_worker"
	RoleCodeCommunityAdmin = "community_admin"
	RoleCodePropertyAdmin  = "property_admin"
	RoleCodeCommittee      = "committee"
	RoleCodeMerchant       = "merchant"
	// RoleCodeRegisteredUser 注册即自动分配的基角色（permission-service sys_role id=9，browse-only、空数据范围、永久有效）
	RoleCodeRegisteredUser = "registered_user"
)

// ==================== 权限 scope 类型（permission-service rel_user_role scope_type） ====================

const (
	ScopeTypeGlobal    = "global"    // 全局范围（审核员 / sys_admin / moderation 系统身份）
	ScopeTypeEmpty     = ""          // 空范围（仅 registered_user 基角色，对任何查询零贡献）
	ScopeTypeCommunity = "community" // 限定节点（小区，scope_id=community_id）
)

// ==================== 最大小区数 ====================

// MaxCommunities 最多加入 3 个小区
// TODO: 可通过 sysconfig 动态配置（key: user.max_community_join_count）
const MaxCommunities = 3

// ==================== 频次限制 ====================

// MaxNewCommunitiesPerYear 每年最多首次加入 3 个新小区
// TODO: 可通过 sysconfig 动态配置（key: user.max_new_communities_per_year）
const MaxNewCommunitiesPerYear = 3

// MaxTotalCommunitiesLifetime 终身最多首次加入 12 个不同小区
// TODO: 可通过 sysconfig 动态配置（key: user.max_total_communities_lifetime）
const MaxTotalCommunitiesLifetime = 12

// MaxHouseMembers 同一房屋（同小区同楼/单元/房号）最多活跃成员数
// TODO: 可通过 sysconfig 动态配置（key: user.max_house_members）
const MaxHouseMembers = 6

// ==================== 角色是否需要房屋 ====================

var RolesRequiringResidence = map[string]bool{
	RoleCodeOwner:  true,
	RoleCodeTenant: true,
}

// ==================== 角色过期策略：true = 有时限 ====================

var RolesWithExpiry = map[string]bool{
	RoleCodeTenant:         true,
	RoleCodeGridWorker:     true,
	RoleCodeCommunityAdmin: true,
	RoleCodePropertyAdmin:  true,
	RoleCodeCommittee:      true,
}

// ==================== 角色默认有效期 ====================

const (
	DefaultGridWorkerExpiryHours     = 8760  // 1 年 = 365 * 24
	DefaultCommunityAdminExpiryHours = 17520 // 2 年
	DefaultPropertyAdminExpiryHours  = 8760  // 1 年
	DefaultCommitteeExpiryHours      = 17520 // 2 年
)

// UnixTimeOrZero converts a sql.NullTime to unix seconds (0 if invalid).
func UnixTimeOrZero(t sql.NullTime) int64 {
	if t.Valid {
		return t.Time.Unix()
	}
	return 0
}

// TimePtr returns a sql.NullTime from a time.Time pointer.
func TimePtr(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}
