package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ListPermissions 查询权限树（树形结构，按 sort_order 排序）
func (l *ListPermissionsLogic) ListPermissions(req *types.ListPermissionsReq) (*types.ListPermissionsResp, error) {
	grpcReq := &permissionv1.ListPermissionsRequest{}
	if req.Type != nil {
		grpcReq.Type = req.Type
	}
	if req.Status != nil {
		grpcReq.Status = req.Status
	}

	grpcResp, err := l.svcCtx.PermissionRpc.ListPermissions(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("ListPermissions gRPC call failed: %v", err)
		return nil, err
	}

	// Base 检查：业务错误写入 grpcResp.Base，需转为 Go error，禁止 deref Permissions
	// SEE: [[rpc-callback-must-check-response-base]]
	if err := responsex.ToError(grpcResp.Base); err != nil {
		return nil, err
	}

	return &types.ListPermissionsResp{
		Permissions: toPermissionInfoList(grpcResp.Permissions),
	}, nil
}
