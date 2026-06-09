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
)

// ==================== 最大小区数 ====================

const MaxCommunities = 3

// ==================== 频次限制 ====================

const MaxNewCommunitiesPerYear = 3    // 每年最多首次加入 3 个新小区
const MaxTotalCommunitiesLifetime = 12 // 终身最多首次加入 12 个不同小区

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
