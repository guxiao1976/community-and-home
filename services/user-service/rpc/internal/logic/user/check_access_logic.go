package user

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckAccessLogic {
	return &CheckAccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckAccess 权限校验（供 API Gateway 实时调用）
// 从 permission-service 获取用户已认证角色（status=2），检查是否匹配指定角色
//
// 使用场景：
//   - Gateway 对敏感操作（审核认证、发通知）做实时鉴权
//   - 规则：role_codes 任一匹配 + community_id 匹配（或 global）+ 角色已认证(status=2)
func (l *CheckAccessLogic) CheckAccess(in *userv1.CheckAccessRequest) (*userv1.CheckAccessResponse, error) {
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("CheckAccess: PermissionClient is nil")
		return &userv1.CheckAccessResponse{
			Base:    responsex.NewBaseRespWithError(50000, "系统繁忙"),
			Allowed: false,
		}, nil
	}

	// 调 permission-service 获取用户所有角色（含生命周期状态）
	resp, err := l.svcCtx.PermissionClient.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: in.UserId})
	if err != nil {
		l.Errorf("GetUserRoles from permission-service failed: %v", err)
		return nil, err
	}

	// 匹配：role_code 命中 + 已认证(status=2) + scope 匹配
	roleSet := make(map[string]bool)
	for _, rc := range in.RoleCodes {
		roleSet[rc] = true
	}

	for _, r := range resp.Roles {
		if r.Status != 2 {
			continue // 只认已认证角色
		}
		if !roleSet[r.Role.Code] {
			continue // 角色不匹配
		}
		// scope 匹配：community 需匹配 communityId；global 放行
		if r.ScopeType == "global" || r.ScopeId == in.CommunityId {
			return &userv1.CheckAccessResponse{
				Base:               responsex.NewBaseResp(),
				Allowed:            true,
				MatchedRole:        r.Role.Code,
				MatchedCommunityId: r.ScopeId,
			}, nil
		}
	}

	return &userv1.CheckAccessResponse{
		Base:    responsex.NewBaseResp(),
		Allowed: false,
	}, nil
}
