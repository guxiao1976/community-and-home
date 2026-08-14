package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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

// CreateRole 创建角色（支持同时分配权限 + 允许登录端）
func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleReq) (*types.CreateRoleResp, error) {
	grpcReq := &permissionv1.CreateRoleRequest{
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		SortOrder:     req.SortOrder,
		PermissionIds: req.PermissionIds,
		Platforms:     req.Platforms, // REQ-PLAT-2 透传
	}

	grpcResp, err := l.svcCtx.PermissionRpc.CreateRole(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("CreateRole gRPC call failed: %v", err)
		return nil, err
	}

	// Base 检查：业务错误（如 60006 角色编码已存在 / 60008 非法登录端）写入 grpcResp.Base，需转为 Go error，
	// 禁止 deref grpcResp.Role（REQ-UPDATE-3，防空指针 panic）
	// SEE: [[rpc-callback-must-check-response-base]]
	if err := responsex.ToError(grpcResp.Base); err != nil {
		return nil, err
	}

	return &types.CreateRoleResp{
		Role: toRoleInfo(grpcResp.Role),
	}, nil
}
