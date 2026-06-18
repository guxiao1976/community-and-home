package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRolesByIdsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRolesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolesByIdsLogic {
	return &GetRolesByIdsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetRolesByIds 根据 ID 列表批量查询角色（Internal 接口，不含权限）
func (l *GetRolesByIdsLogic) GetRolesByIds(in *permissionv1.GetRolesByIdsRequest) (*permissionv1.GetRolesByIdsResponse, error) {
	if len(in.Ids) == 0 {
		return &permissionv1.GetRolesByIdsResponse{
			Base:  responsex.NewBaseResp(),
			Roles: nil,
		}, nil
	}

	roles, err := l.svcCtx.RoleModel.FindByIds(l.ctx, in.Ids)
	if err != nil {
		return nil, err
	}

	var pbRoles []*permissionv1.Role
	for _, r := range roles {
		pbRoles = append(pbRoles, toRolePb(r))
	}

	return &permissionv1.GetRolesByIdsResponse{
		Base:  responsex.NewBaseResp(),
		Roles: pbRoles,
	}, nil
}
