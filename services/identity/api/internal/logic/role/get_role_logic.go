// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get role details
func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleLogic) GetRole(req *types.GetRoleReq) (resp *types.GetRoleResp, err error) {
	// Get role
	role, err := l.svcCtx.AuthRoleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			logx.Errorf("Role not found: %d", req.Id)
			return nil, err
		}
		logx.Errorf("Failed to get role: %v", err)
		return nil, err
	}

	// Get role permissions
	rolePermissions, err := l.svcCtx.AuthRolePermissionModel.FindByRoleId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("Failed to get role permissions: %v", err)
		return nil, err
	}

	// Get permission details
	var permissions []types.Permission
	if len(rolePermissions) > 0 {
		permissionIds := make([]int64, len(rolePermissions))
		for i, rp := range rolePermissions {
			permissionIds[i] = rp.PermissionId
		}

		permissionList, err := l.svcCtx.AuthPermissionModel.FindByIds(l.ctx, permissionIds)
		if err != nil {
			logx.Errorf("Failed to get permissions: %v", err)
			return nil, err
		}

		permissions = make([]types.Permission, 0, len(permissionList))
		for _, p := range permissionList {
			permissions = append(permissions, types.Permission{
				Id:          p.Id,
				Name:        p.Name,
				Code:        p.Code,
				Type:        int32(p.Type),
				ParentId:    &p.ParentId.Int64,
				Path:        p.Path.String,
				SortOrder:   int32(p.SortOrder),
				Status:      int32(p.Status),
				CreatedTime: p.CreatedTime.Format("2006-01-02 15:04:05"),
				UpdatedTime: p.UpdatedTime.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return &types.GetRoleResp{
		Role: types.RoleWithPermissions{
			Id:          role.Id,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description.String,
			IsSystem:    int32(role.IsSystem),
			Status:      int32(role.Status),
			Permissions: permissions,
			CreatedTime: role.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: role.UpdatedTime.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

