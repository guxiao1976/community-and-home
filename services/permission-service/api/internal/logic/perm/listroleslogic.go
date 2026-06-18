package perm

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ListRoles 分页查询角色列表
func (l *ListRolesLogic) ListRoles(req *types.ListRolesReq) (*types.ListRolesResp, error) {
	// 构建 gRPC 请求
	grpcReq := &permissionv1.ListRolesRequest{
		Page: &commonv1.PageRequest{
			Page:     int32(req.Page),
			PageSize: int32(req.PageSize),
		},
	}
	if req.Status != nil {
		grpcReq.Status = req.Status
	}

	// 调用 gRPC
	grpcResp, err := l.svcCtx.PermissionRpc.ListRoles(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("ListRoles gRPC call failed: %v", err)
		return nil, err
	}

	// 转换响应
	roles := make([]types.RoleInfo, 0, len(grpcResp.Roles))
	for _, r := range grpcResp.Roles {
		roles = append(roles, toRoleInfo(r))
	}

	resp := &types.ListRolesResp{
		Roles: roles,
	}
	if grpcResp.Page != nil {
		resp.Page = types.PageInfo{
			Page:       grpcResp.Page.Page,
			PageSize:   grpcResp.Page.PageSize,
			Total:      grpcResp.Page.Total,
			TotalPages: grpcResp.Page.TotalPages,
		}
	}

	return resp, nil
}
