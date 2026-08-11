package user

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserRoles 查询用户角色（从 permission-service 获取）
//
// 缓存策略：
//   - 仅已认证角色（status=2）查询走缓存（auth-service 签发 JWT 高频调用）
//   - 其他查询直接调 permission-service（低频）
//   - TTL 300s，角色状态变更时 permission-service 主动失效
func (l *GetUserRolesLogic) GetUserRoles(in *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("GetUserRoles: PermissionClient is nil")
		return &userv1.GetUserRolesResponse{
			Base: responsex.NewBaseRespWithError(50000, "系统繁忙"),
		}, nil
	}

	// Cache-Aside: 仅已认证角色查询走缓存
	onlyApproved := in.VerfStatus != nil && *in.VerfStatus == model.RoleVerfStatusApproved
	if onlyApproved {
		if cached := getRolesFromCache(l.ctx, l.svcCtx.Redis, in.UserId); cached != nil {
			return cached, nil
		}
	}

	// 调 permission-service 获取所有角色（含生命周期状态）
	resp, err := l.svcCtx.PermissionClient.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: in.UserId})
	if err != nil {
		l.Errorf("GetUserRoles from permission-service failed: %v", err)
		return nil, err
	}

	// 转换 proto（permission UserRoleInfo → user MembershipRole）
	var roles []*userv1.MembershipRole
	for _, r := range resp.Roles {
		// 过滤：仅已认证时只返回 status=2
		if onlyApproved && r.Status != 2 {
			continue
		}
		mr := &userv1.MembershipRole{
			UserId:      in.UserId,
			RoleCode:    r.Role.Code,
			CommunityId: r.ScopeId,
			VerfStatus:  r.Status,
			ExpiresAt:   r.ExpiresAt,
		}
		if r.VerifiedAt > 0 {
			mr.VerifiedAt = r.VerifiedAt
		}
		roles = append(roles, mr)
	}

	result := &userv1.GetUserRolesResponse{
		Base:  responsex.NewBaseResp(),
		Roles: roles,
	}

	// 回填缓存（仅已认证角色）
	if onlyApproved {
		ttl := 300
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.cache.roles_ttl_seconds"); err == nil {
				ttl = v
			}
		}
		setRolesToCache(l.ctx, l.svcCtx.Redis, in.UserId, result, ttl)
	}

	return result, nil
}
