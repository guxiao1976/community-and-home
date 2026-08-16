package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPublishPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishPermissionLogic {
	return &GetPublishPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// publishRoleCodeToContentPostRole RBAC code → ContentPostRole 映射（D6，property_admin 保留）。
func publishRoleCodeToContentPostRole(code string) (communityv1.ContentPostRole, bool) {
	switch code {
	case scope.RoleGridWorker:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_GRID_OFFICER, true
	case scope.RoleCommunityAdmin:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMUNITY, true
	case scope.RoleCommittee:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMITTEE, true
	case scope.RolePropertyAdmin:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_PROPERTY, true
	default:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_UNSPECIFIED, false
	}
}

// GetPublishPermission can_publish + 可发布角色（含 property_admin）。
//
// level-2 判定（D6）：role.Code ∈ {grid_worker, community_admin, property_admin, committee} 且
// status==2 且 verified_at>0 且未过期（基于 RPC 输出，禁止直读 rel_user_role）。
// owner/tenant/merchant/sys_admin → can_publish=false。
func (l *GetPublishPermissionLogic) GetPublishPermission(in *communityv1.GetPublishPermissionRequest) (*communityv1.GetPublishPermissionResponse, error) {
	userID := scope.UserIDFromCtx(l.ctx)
	if userID == 0 {
		// 防御：认证中间件兜底（未登录由认证中间件 UNAUTHENTICATED，此处视为 false）
		return &communityv1.GetPublishPermissionResponse{
			Base:       responsex.NewBaseResp(),
			CanPublish: false,
		}, nil
	}

	resp, err := l.svcCtx.PermissionClient.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: userID})
	if err != nil {
		l.Errorf("GetPublishPermission: get user roles failed: %v", err)
		return nil, err
	}

	canPublish := false
	roleSet := make(map[communityv1.ContentPostRole]bool)
	for _, ur := range resp.GetRoles() {
		if !scope.IsLevel2Grant(ur) {
			continue
		}
		if pr, ok := publishRoleCodeToContentPostRole(ur.GetRole().GetCode()); ok {
			canPublish = true
			roleSet[pr] = true
		}
	}

	// publishable_roles 固定优先序（与 PublishRolesFrom 同序：grid_worker > community_admin > committee > property_admin）
	var publishable []communityv1.ContentPostRole
	if canPublish {
		order := []communityv1.ContentPostRole{
			communityv1.ContentPostRole_CONTENT_POST_ROLE_GRID_OFFICER,
			communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMUNITY,
			communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMITTEE,
			communityv1.ContentPostRole_CONTENT_POST_ROLE_PROPERTY,
		}
		publishable = make([]communityv1.ContentPostRole, 0, len(order))
		for _, r := range order {
			if roleSet[r] {
				publishable = append(publishable, r)
			}
		}
	}

	return &communityv1.GetPublishPermissionResponse{
		Base:             responsex.NewBaseResp(),
		CanPublish:       canPublish,
		PublishableRoles: publishable,
	}, nil
}
