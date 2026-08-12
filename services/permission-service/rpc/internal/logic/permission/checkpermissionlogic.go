package permission

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// CheckPermission 鉴权检查（T1.5 能力分层重写）
//
//	聚合规则（数据驱动，§5.1.1）：
//	  权限 P 的 min_verf_level = L（0 或 2）
//	  对每个授予 P 的角色 grant g：满足层级(g) ∈ {2, 0, none}（grantSatisfiedLevel）
//	  P 放行 ⟺ max(满足层级(g)) ≥ L
//	缓存：
//	  perm:def:{needle}  — String, min_verf_level（TTL 30min）
//	  perm:user:{userId} — Hash {path: maxLevel}（TTL 30min）
//	所有角色（含 is_system=1）统一走 rel_role_permission 配置，无短路特权
//
// SEE: [[is-system-no-permission-shortcut]]、[[permission-seed-api-path-must-match-routes]]
func (l *CheckPermissionLogic) CheckPermission(in *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	// 0. 检查用户是否被禁用（禁用标记由 user-service 写入）
	disabledKey := fmt.Sprintf("user:disabled:%d", in.UserId)
	disabled, err := l.svcCtx.RedisClient.Exists(l.ctx, disabledKey).Result()
	if err == nil && disabled > 0 {
		l.Infof("CheckPermission denied: user=%d is disabled", in.UserId)
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	// needle: path 字段已包含 Method 前缀（如 "GET:/api/users"），无需再拼接
	needle := in.ApiPath
	if in.Action != "" && !strings.HasPrefix(in.ApiPath, in.Action+":") {
		needle = in.Action + ":" + in.ApiPath
	}

	// 1. 权限定义 min_verf_level（perm:def:{needle} 缓存，MISS 回源 DB）
	minLevel, ok := l.permissionDefMinLevel(needle)
	if !ok {
		l.Infof("CheckPermission denied: permission not found needle=%s user=%d", needle, in.UserId)
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	// 2. 用户满足层级（HGET perm:user:{userId} {needle} → maxLevel）
	maxLevel, ok := l.userMaxLevel(in.UserId, needle)
	if !ok || maxLevel < minLevel {
		l.Infof("CheckPermission denied: user=%d needle=%s maxLevel=%d minVerfLevel=%d", in.UserId, needle, maxLevel, minLevel)
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
}

// permissionDefMinLevel 查权限定义 min_verf_level
//
//	HIT  perm:def:{needle} 返回层级；"-1" sentinel 表示权限不存在
//	MISS 查 sys_permission（先 path 后 code），回填缓存
func (l *CheckPermissionLogic) permissionDefMinLevel(needle string) (int64, bool) {
	defKey := "perm:def:" + needle
	if v, err := l.svcCtx.RedisClient.Get(l.ctx, defKey).Result(); err == nil {
		if v == "-1" {
			return 0, false
		}
		lv, _ := strconv.ParseInt(v, 10, 64)
		return lv, true
	}

	p, err := l.svcCtx.PermissionModel.FindByPath(l.ctx, needle)
	if err != nil {
		p, err = l.svcCtx.PermissionModel.FindByCode(l.ctx, needle)
		if err != nil {
			// 缓存 not-found sentinel，避免反复回源
			l.svcCtx.RedisClient.Set(l.ctx, defKey, "-1", 30*time.Minute)
			return 0, false
		}
	}
	l.svcCtx.RedisClient.Set(l.ctx, defKey, strconv.FormatInt(p.MinVerfLevel, 10), 30*time.Minute)
	return p.MinVerfLevel, true
}

// userMaxLevel 聚合用户对 needle 的满足层级（能力分层）
//
//	HIT  HGET perm:user:{userId} {needle} → maxLevel
//	MISS 查 FindActiveRolesByUserId(status∈{0,1,2}) → 角色→权限 → 逐 path 计算 maxLevel → HSET 回填
func (l *CheckPermissionLogic) userMaxLevel(userId int64, needle string) (int64, bool) {
	permCacheKey := fmt.Sprintf("perm:user:%d", userId)
	if v, err := l.svcCtx.RedisClient.HGet(l.ctx, permCacheKey, needle).Result(); err == nil {
		lv, _ := strconv.ParseInt(v, 10, 64)
		return lv, true
	}

	// 收集用户所有活跃 grants（status∈{0,1,2} 且未过期）
	grants, err := l.svcCtx.UserRoleModel.FindActiveRolesByUserId(l.ctx, userId)
	if err != nil || len(grants) == 0 {
		return 0, false
	}

	// roleId → 权限 ID 列表
	rolePermIds := make(map[int64][]int64, len(grants))
	permIdSet := make(map[int64]struct{})
	for _, g := range grants {
		if _, seen := rolePermIds[g.RoleId]; seen {
			continue
		}
		rps, _ := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, g.RoleId)
		for _, rp := range rps {
			rolePermIds[g.RoleId] = append(rolePermIds[g.RoleId], rp.PermissionId)
			permIdSet[rp.PermissionId] = struct{}{}
		}
	}
	if len(permIdSet) == 0 {
		return 0, false
	}

	permIds := make([]int64, 0, len(permIdSet))
	for pid := range permIdSet {
		permIds = append(permIds, pid)
	}
	perms, err := l.svcCtx.PermissionModel.FindByIds(l.ctx, permIds)
	if err != nil {
		return 0, false
	}
	permById := make(map[int64]*model.SysPermission, len(perms))
	for _, p := range perms {
		permById[p.Id] = p
	}

	// 聚合：path → maxLevel（多角色叠加取最高）
	pathMax := make(map[string]int64)
	for _, g := range grants {
		lv := int64(grantSatisfiedLevel(g))
		if lv < 0 {
			continue
		}
		for _, pid := range rolePermIds[g.RoleId] {
			p, ok := permById[pid]
			if !ok {
				continue
			}
			key := p.Path.String
			if key == "" {
				key = p.Code
			}
			if cur, ok := pathMax[key]; !ok || lv > cur {
				pathMax[key] = lv
			}
		}
	}

	if len(pathMax) > 0 {
		// 回填 Hash 缓存
		for k, lv := range pathMax {
			l.svcCtx.RedisClient.HSet(l.ctx, permCacheKey, k, strconv.FormatInt(lv, 10))
		}
		cacheTTL := 1800
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "permission.cache.ttl_seconds"); err == nil && v > 0 {
				cacheTTL = v
			}
		}
		l.svcCtx.RedisClient.Expire(l.ctx, permCacheKey, time.Duration(cacheTTL)*time.Second)
	}

	lv, ok := pathMax[needle]
	return lv, ok
}
