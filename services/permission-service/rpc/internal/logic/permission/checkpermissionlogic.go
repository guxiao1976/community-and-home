package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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

// CheckPermission 鉴权检查
//
//	组合 User Roles → 查询 API 权限集 → 判断是否包含请求的 API Path
//	先查 Redis 缓存，未命中则查 DB 并回填缓存
//	所有角色（含 is_system=1）统一走 rel_role_permission 配置，无短路特权
func (l *CheckPermissionLogic) CheckPermission(in *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	// 0. 检查用户是否被禁用（禁用标记由 user-service 写入）
	disabledKey := fmt.Sprintf("user:disabled:%d", in.UserId)
	disabled, err := l.svcCtx.RedisClient.Exists(l.ctx, disabledKey).Result()
	if err == nil && disabled > 0 {
		l.Infof("CheckPermission denied: user=%d is disabled", in.UserId)
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	permCacheKey := fmt.Sprintf("perm:user:%d", in.UserId)
	// needle: path 字段已包含 Method 前缀（如 "GET:/api/users"），无需再拼接
	needle := in.ApiPath
	if in.Action != "" && !strings.HasPrefix(in.ApiPath, in.Action+":") {
		needle = in.Action + ":" + in.ApiPath
	}

	// 1. 查 Redis 缓存
	cached, err := l.svcCtx.RedisClient.SIsMember(l.ctx, permCacheKey, needle).Result()
	if err == nil && cached {
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
	}

	// 2. 缓存未命中 → 查 DB
	roles, err := l.svcCtx.UserRoleModel.FindActiveByUserId(l.ctx, in.UserId)
	if err != nil || len(roles) == 0 {
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	// 收集 role_ids → permissions
	roleIds := make([]int64, len(roles))
	for i, r := range roles {
		roleIds[i] = r.RoleId
	}

	// 查 role-permission 关联
	var permIds []int64
	for _, rid := range roleIds {
		rps, _ := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, rid)
		for _, rp := range rps {
			permIds = append(permIds, rp.PermissionId)
		}
	}
	if len(permIds) == 0 {
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	perms, err := l.svcCtx.PermissionModel.FindByIds(l.ctx, permIds)
	if err != nil {
		return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
	}

	// 匹配 + 回填 Redis
	for _, p := range perms {
		// path 字段已包含 Method 前缀（如 "GET:/api/users"），直接使用
		cacheKey := p.Path.String
		if cacheKey == "" {
			cacheKey = fmt.Sprintf("%s:%s", in.Action, p.Code)
		}
		l.svcCtx.RedisClient.SAdd(l.ctx, permCacheKey, cacheKey)
		if cacheKey == needle || p.Code == needle {
			return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		}
	}
	// 设置缓存 TTL（可配置，默认 30 分钟）
	cacheTTL := 1800
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "permission.cache.ttl_seconds"); err == nil && v > 0 {
			cacheTTL = v
		}
	}
	l.svcCtx.RedisClient.Expire(l.ctx, permCacheKey, time.Duration(cacheTTL)*time.Second)

	return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: false}, nil
}
