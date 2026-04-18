// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package permission

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get permission tree
func NewGetPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionsLogic {
	return &GetPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionsLogic) GetPermissions(req *types.GetPermissionsReq) (resp *types.GetPermissionsResp, err error) {
	// Convert status to int64 pointer
	var status *int64
	if req.Status != nil {
		s := int64(*req.Status)
		status = &s
	}

	// Get all permissions
	permissions, err := l.svcCtx.AuthPermissionModel.FindAll(l.ctx, status)
	if err != nil {
		logx.Errorf("Failed to get permissions: %v", err)
		return nil, err
	}

	// Build permission tree
	tree := BuildPermissionTree(permissions)

	return &types.GetPermissionsResp{
		List: tree,
	}, nil
}

