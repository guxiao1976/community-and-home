package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetUserRoles 查询用户拥有的所有角色（含作用域信息）
//   联表 JOIN rel_user_role + sys_role
func (l *GetUserRolesLogic) GetUserRoles(in *permissionv1.GetUserRolesRequest) (*permissionv1.GetUserRolesResponse, error) {
	roles, err := l.svcCtx.UserRoleModel.FindActiveByUserId(l.ctx, in.UserId)
	if err != nil || len(roles) == 0 {
		return &permissionv1.GetUserRolesResponse{
			Base:  responsex.NewBaseResp(),
			Roles: nil,
		}, nil
	}

	var pbRoles []*permissionv1.UserRoleInfo
	for _, r := range roles {
		pbRoles = append(pbRoles, &permissionv1.UserRoleInfo{
			Role: &permissionv1.Role{
				Id:          r.RoleId,
				Code:        r.RoleCode,
				Name:        r.RoleName,
				Description: r.Description,
				IsSystem:    r.IsSystem == 1,
				Status:      int32(r.Status),
			},
			ScopeType: r.ScopeType,
			ScopeId:   r.ScopeId,
		})
	}

	return &permissionv1.GetUserRolesResponse{
		Base:  responsex.NewBaseResp(),
		Roles: pbRoles,
	}, nil
}
