package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// UpdateRole 更新角色信息
func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleReq) (*types.UpdateRoleResp, error) {
	grpcReq := &permissionv1.UpdateRoleRequest{
		Id:            req.Id,
		PermissionIds: req.PermissionIds,
	}
	if req.Name != nil {
		grpcReq.Name = req.Name
	}
	if req.Description != nil {
		grpcReq.Description = req.Description
	}
	if req.Status != nil {
		grpcReq.Status = req.Status
	}
	if req.SortOrder != nil {
		grpcReq.SortOrder = req.SortOrder
	}

	grpcResp, err := l.svcCtx.PermissionRpc.UpdateRole(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("UpdateRole gRPC call failed: %v", err)
		return nil, err
	}

	return &types.UpdateRoleResp{
		Role: toRoleInfo(grpcResp.Role),
	}, nil
}
