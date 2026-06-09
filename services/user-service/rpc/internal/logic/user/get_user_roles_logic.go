package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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

// GetUserRoles 查询用户角色（含 Redis Cache-Aside）
//
// 缓存策略：
//   - 仅 verf_status=2（已认证）走缓存（auth-service 签发 JWT 高频调用）
//   - 其他查询直接查 DB（低频，不需要缓存）
//   - TTL 300s，角色状态变更时主动失效
func (l *GetUserRolesLogic) GetUserRoles(in *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	// Cache-Aside: 仅已认证角色查询走缓存
	if in.VerfStatus != nil && *in.VerfStatus == model.RoleVerfStatusApproved {
		if cached := getRolesFromCache(l.ctx, l.svcCtx.Redis, in.UserId); cached != nil {
			return cached, nil
		}
	}

	// 查询 DB
	var roles []*model.UserMembershipRole
	var err error

	if in.VerfStatus != nil {
		roles, err = l.svcCtx.UserMembershipRoleModel.FindApprovedByUser(
			l.ctx, in.UserId, in.CommunityId, nil)
	} else if in.CommunityId > 0 {
		roles, err = l.svcCtx.UserMembershipRoleModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
	} else {
		roles, err = l.svcCtx.UserMembershipRoleModel.FindByUserId(l.ctx, in.UserId)
	}

	if err != nil {
		l.Errorf("find roles error: %v", err)
		return nil, err
	}

	resp := &userv1.GetUserRolesResponse{
		Base:  responsex.NewBaseResp(),
		Roles: toProtoRoles(roles),
	}

	// 回填缓存（仅已认证角色）
	if in.VerfStatus != nil && *in.VerfStatus == model.RoleVerfStatusApproved {
		ttl := 300
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.cache.roles_ttl_seconds"); err == nil {
				ttl = v
			}
		}
		setRolesToCache(l.ctx, l.svcCtx.Redis, in.UserId, resp, ttl)
	}

	return resp, nil
}
