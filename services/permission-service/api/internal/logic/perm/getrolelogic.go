package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetRole 获取单个角色详情（含权限列表）
func (l *GetRoleLogic) GetRole(req *types.GetRoleReq) (*types.GetRoleResp, error) {
	grpcReq := &permissionv1.GetRoleRequest{
		Id: req.Id,
	}

	grpcResp, err := l.svcCtx.PermissionRpc.GetRole(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("GetRole gRPC call failed: %v", err)
		return nil, err
	}

	return &types.GetRoleResp{
		Role: toRoleInfo(grpcResp.Role),
	}, nil
}
