package permission

import (
	"context"
	"database/sql"
	"strings"
	"time"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-permission/model"
)

// CodeInvalidPlatform 非法登录端错误码（060008，协议头注释登记）
// SEE: [[error-code-literal-bypasses-qa-gate]] — 禁止裸字面量，必须命名常量
const CodeInvalidPlatform int32 = 60008

// CodeCommunityAdminLimit 每小区 community_admin 人数上限错误码（060009，协议头注释待 Owner 同步）
// 社区管理员是数据范围特权角色（驱动 division 子树发布范围展开），人数必须受控。
// SEE: [[error-code-collision-and-namespace-alignment]] — 60001-60008 已占（含文档登记的 60002/60003），
// 新增语义必须分配新码段 60009（一码一义，禁止复用 60002「权限不存在」）
const CodeCommunityAdminLimit int32 = 60009

// grantSatisfiedLevel 计算单个 grant 满足的能力层级（T1.5 §5.1.1 数据驱动聚合规则）
//
//	level-2 = status==2 AND verified_at NOT NULL AND 未过期
//	level-0 = status∈{0,1} AND 未过期；或 status==2 AND verified_at NULL（registered_user 恒 level-0，S2）
//	status∈{3,4} / 已过期 → 不计（-1）
//
// 返回 -1 表示该 grant 不满足任何层级（不进并集）。
func grantSatisfiedLevel(g *model.UserRoleWithInfo) int {
	if !grantActive(g) {
		return -1
	}
	// 过期防御（SQL 已过滤，这里双保险）
	if g.ExpiresAt.Valid && !g.ExpiresAt.Time.After(time.Now()) {
		return -1
	}
	if g.URStatus == 2 && g.VerifiedAt.Valid {
		return 2
	}
	return 0
}

// scopeCacheData 读穿缓存 JSON 结构（key: perm:scopes:{userId}:{scopeType}）
// 设计 §4.1：{"state":"empty|global|limited","ids":[int64]}
type scopeCacheData struct {
	State string  `json:"state"`
	Ids   []int64 `json:"ids"`
}

// scopeStateString 将三态枚举映射为缓存字符串（empty|limited|global）
func scopeStateString(s permissionv1.DataScopeState) string {
	switch s {
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL:
		return "global"
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED:
		return "limited"
	default:
		return "empty"
	}
}

// scopeStateFromString 将缓存字符串映射回三态枚举（未知/空串 → EMPTY，安全默认）
func scopeStateFromString(s string) permissionv1.DataScopeState {
	switch s {
	case "global":
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL
	case "limited":
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED
	default:
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY
	}
}

// sqlNullString converts a string to sql.NullString (empty string → NULL)
func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// splitPlatforms 将逗号分隔的 platforms 字符串切分为切片（空串 → 空切片，非 nil）
//
//	platforms 为配置属性（pc/mobile），与权限 path 匹配无关，仅透出给 auth-service 端准入判定。
//	SEE: [[is-system-no-permission-shortcut]] — platforms 不参与权限短路，空值运行时 fail-open
func splitPlatforms(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, raw := range parts {
		if p := strings.TrimSpace(raw); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// joinPlatforms 将切片合并为逗号分隔字符串（写回存储时使用，与 splitPlatforms 互为逆操作）
func joinPlatforms(ps []string) string {
	return strings.Join(ps, ",")
}

// validPlatforms 允许登录端值域（pc/mobile），与前端 PLATFORM_OPTIONS 对齐
// 平台值域权威在后端：非法端由 RPC 60008 拒绝
// SEE: [[frontend-business-rule-hardcode]] — 后端为值域唯一权威
var validPlatforms = map[string]struct{}{
	"pc":     {},
	"mobile": {},
}

// validatePlatforms 校验 platforms 值域并去重（REQ-PLAT-4）
//
//   - 值域 {pc, mobile}；任一非法值 → errx.CodeError(60008, "非法登录端: <v>")
//   - 空/nil → 通过（fail-open，允许所有端）
//   - 重复值 → 去重且保持原顺序（["pc","pc","mobile"] → ["pc","mobile"]）
func validatePlatforms(ps []string) ([]string, error) {
	out := make([]string, 0, len(ps))
	seen := make(map[string]struct{}, len(ps))
	for _, v := range ps {
		if _, ok := validPlatforms[v]; !ok {
			return nil, errx.NewCodeError(int(CodeInvalidPlatform), "非法登录端: "+v)
		}
		if _, dup := seen[v]; dup {
			continue // 去重，保持首次出现顺序
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out, nil
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
		Platforms:   splitPlatforms(r.Platforms),
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
			Timestamps: &commonv1.Timestamps{
				CreatedAt: p.CreatedTime.Unix(),
				UpdatedAt: p.UpdatedTime.Unix(),
			},
		})
	}
	pb.Permissions = pbPerms
	return pb
}
