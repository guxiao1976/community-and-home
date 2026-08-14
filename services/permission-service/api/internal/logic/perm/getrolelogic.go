package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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

	// Base 检查：业务错误（如 60001 角色不存在）写入 grpcResp.Base，需转为 Go error，禁止 deref grpcResp.Role
	// SEE: [[rpc-callback-must-check-response-base]]
	if err := responsex.ToError(grpcResp.Base); err != nil {
		return nil, err
	}

	return &types.GetRoleResp{
		Role: toRoleInfo(grpcResp.Role),
	}, nil
}
