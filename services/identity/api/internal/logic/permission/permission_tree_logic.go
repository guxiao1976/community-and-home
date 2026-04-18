package permission

import (
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
)

// BuildPermissionTree builds a tree structure from flat permission list
func BuildPermissionTree(permissions []*model.AuthPermission) []types.PermissionTree {
	// Create a map for quick lookup
	permissionMap := make(map[int64]*types.PermissionTree)
	var rootPermissions []types.PermissionTree

	// First pass: create all nodes
	for _, p := range permissions {
		node := types.PermissionTree{
			Id:          p.Id,
			Name:        p.Name,
			Code:        p.Code,
			Type:        int32(p.Type),
			ParentId:    nil,
			Path:        p.Path.String,
			SortOrder:   int32(p.SortOrder),
			Status:      int32(p.Status),
			Children:    []types.PermissionTree{},
			CreatedTime: p.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: p.UpdatedTime.Format("2006-01-02 15:04:05"),
		}

		if p.ParentId.Valid {
			parentId := p.ParentId.Int64
			node.ParentId = &parentId
		}

		permissionMap[p.Id] = &node
	}

	// Second pass: build tree structure
	for _, p := range permissions {
		node := permissionMap[p.Id]
		if p.ParentId.Valid {
			// Has parent, add to parent's children
			if parent, exists := permissionMap[p.ParentId.Int64]; exists {
				parent.Children = append(parent.Children, *node)
			}
		} else {
			// Root node
			rootPermissions = append(rootPermissions, *node)
		}
	}

	return rootPermissions
}
