package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolePermissionsLogic {
	return &GetRolePermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetRolePermissions 获取角色已分配的权限 ID 列表
func (l *GetRolePermissionsLogic) GetRolePermissions(req *types.GetRoleReq) (*types.GetRolePermissionsResp, error) {
	grpcResp, err := l.svcCtx.PermissionRpc.GetRole(l.ctx, &permissionv1.GetRoleRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}

	// Base 检查：60001 → 转 Go error，禁止 deref grpcResp.Role.Permissions（否则 nil 指针 panic）
	// SEE: [[rpc-callback-must-check-response-base]]
	if err := responsex.ToError(grpcResp.Base); err != nil {
		return nil, err
	}

	var permIds []int64
	for _, p := range grpcResp.Role.Permissions {
		permIds = append(permIds, p.Id)
	}

	return &types.GetRolePermissionsResp{
		PermissionIds: permIds,
	}, nil
}
