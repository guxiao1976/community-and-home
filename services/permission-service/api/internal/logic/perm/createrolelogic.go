package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// CreateRole 创建角色（支持同时分配权限）
func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleReq) (*types.CreateRoleResp, error) {
	grpcReq := &permissionv1.CreateRoleRequest{
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		SortOrder:     req.SortOrder,
		PermissionIds: req.PermissionIds,
	}

	grpcResp, err := l.svcCtx.PermissionRpc.CreateRole(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("CreateRole gRPC call failed: %v", err)
		return nil, err
	}

	return &types.CreateRoleResp{
		Role: toRoleInfo(grpcResp.Role),
	}, nil
}
