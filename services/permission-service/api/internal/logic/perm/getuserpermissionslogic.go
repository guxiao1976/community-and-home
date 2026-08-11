package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetUserPermissions 查询指定用户的权限编码集合（管理员操作）
func (l *GetUserPermissionsLogic) GetUserPermissions(req *types.GetUserPermissionsReq) (*types.GetUserPermissionsResp, error) {
	grpcResp, err := l.svcCtx.PermissionRpc.GetUserPermissions(l.ctx, &permissionv1.GetUserPermissionsRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}

	return &types.GetUserPermissionsResp{
		PermissionCodes: grpcResp.PermissionCodes,
	}, nil
}
