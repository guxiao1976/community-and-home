package permission

import (
	"context"
	"fmt"
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

// CheckPermission 鉴权检查（spec/permission.md 核心逻辑流 3）
//   组合 User Roles → 查询 API 权限集 → 判断是否包含请求的 API Path
//   系统角色（is_system=1）直接 allowed=true
//   先查 Redis 缓存，未命中则查 DB 并回填缓存
func (l *CheckPermissionLogic) CheckPermission(in *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	permCacheKey := fmt.Sprintf("perm:user:%d", in.UserId)
	needle := fmt.Sprintf("%s:%s", in.Action, in.ApiPath)

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

	// 系统角色直接放行
	for _, r := range roles {
		if r.IsSystem == 1 {
			return &permissionv1.CheckPermissionResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		}
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
		code := fmt.Sprintf("%s:%s", in.Action, p.Path.String)
		l.svcCtx.RedisClient.SAdd(l.ctx, permCacheKey, code)
		if code == needle || p.Code == needle {
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
