// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserPermissionsLogic) GetUserPermissions(req *types.GetUserPermissionsReq) (resp *types.GetUserPermissionsResp, err error) {
	l.Infof("GetUserPermissions called for userId: %d", req.UserId)

	// 1. Get user's active roles
	userRoles, err := l.svcCtx.AuthUserRoleModel.FindActiveByUserId(l.ctx, req.UserId)
	if err != nil {
		l.Errorf("FindActiveByUserId error: %v", err)
		if err == sql.ErrNoRows || err == model.ErrNotFound {
			return &types.GetUserPermissionsResp{
				Permissions: []string{},
				Menus:       []types.PermissionTree{},
			}, nil
		}
		return nil, err
	}

	l.Infof("Found %d user roles", len(userRoles))
	if len(userRoles) == 0 {
		return &types.GetUserPermissionsResp{
			Permissions: []string{},
			Menus:       []types.PermissionTree{},
		}, nil
	}

	// 2. Collect all permission IDs from all roles
	permissionIds := make(map[int64]struct{})
	for _, ur := range userRoles {
		l.Infof("Processing role_id: %d", ur.RoleId)
		rp, err := l.svcCtx.AuthRolePermissionModel.FindByRoleId(l.ctx, ur.RoleId)
		if err != nil {
			l.Errorf("FindByRoleId error for role_id %d: %v", ur.RoleId, err)
			return nil, err
		}
		l.Infof("Found %d permissions for role_id: %d", len(rp), ur.RoleId)
		for _, p := range rp {
			permissionIds[p.PermissionId] = struct{}{}
		}
	}

	l.Infof("Total unique permission IDs: %d", len(permissionIds))
	if len(permissionIds) == 0 {
		return &types.GetUserPermissionsResp{
			Permissions: []string{},
			Menus:       []types.PermissionTree{},
		}, nil
	}

	// 3. Fetch all permissions by IDs
	ids := make([]int64, 0, len(permissionIds))
	for id := range permissionIds {
		ids = append(ids, id)
	}

	l.Infof("Fetching permissions for IDs: %v", ids)
	permissions, err := l.svcCtx.AuthPermissionModel.FindByIds(l.ctx, ids)
	if err != nil {
		l.Errorf("FindByIds error: %v", err)
		return nil, err
	}
	l.Infof("Found %d permissions", len(permissions))

	// 4. Build response: permission codes (active only) and menu tree
	var permissionCodes []string
	for _, p := range permissions {
		if p.Status == 1 {
			permissionCodes = append(permissionCodes, p.Code)
		}
	}

	l.Infof("Active permission codes: %v", permissionCodes)
	menus := buildPermissionTree(permissions)

	return &types.GetUserPermissionsResp{
		Permissions: permissionCodes,
		Menus:       menus,
	}, nil
}

func buildPermissionTree(permissions []*model.AuthPermission) []types.PermissionTree {
	// Only include active permissions
	activeMap := make(map[int64]*model.AuthPermission)
	for _, p := range permissions {
		if p.Status == 1 {
			activeMap[p.Id] = p
		}
	}

	treeMap := make(map[int64]*types.PermissionTree)
	var roots []*types.PermissionTree

	for _, p := range activeMap {
		path := ""
		if p.Path.Valid {
			path = p.Path.String
		}
		node := &types.PermissionTree{
			Id:          p.Id,
			Name:        p.Name,
			Code:        p.Code,
			Type:        int32(p.Type),
			Path:        path,
			SortOrder:   int32(p.SortOrder),
			Status:      int32(p.Status),
			Children:    []types.PermissionTree{},
			CreatedTime: p.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: p.UpdatedTime.Format("2006-01-02 15:04:05"),
		}
		if p.ParentId.Valid {
			node.ParentId = &p.ParentId.Int64
		}
		treeMap[p.Id] = node
	}

	for _, node := range treeMap {
		if node.ParentId != nil {
			if parent, ok := treeMap[*node.ParentId]; ok {
				parent.Children = append(parent.Children, *node)
				continue
			}
		}
		roots = append(roots, node)
	}

	if len(roots) == 0 {
		return []types.PermissionTree{}
	}
	result := make([]types.PermissionTree, len(roots))
	for i, r := range roots {
		result[i] = *r
	}
	return result
}
