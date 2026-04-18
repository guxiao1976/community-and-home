package role

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// SyncRolePolicies syncs role permissions to Casbin
func SyncRolePolicies(ctx context.Context, enforcer *casbin.Enforcer, roleId int64, permissionCodes []string) error {
	roleCode := fmt.Sprintf("role:%d", roleId)

	// Remove all existing policies for this role
	_, err := enforcer.RemoveFilteredPolicy(0, roleCode)
	if err != nil {
		logx.Errorf("Failed to remove existing policies for role %d: %v", roleId, err)
		return err
	}

	// Add new policies
	for _, permCode := range permissionCodes {
		_, err := enforcer.AddPolicy(roleCode, permCode, "allow")
		if err != nil {
			logx.Errorf("Failed to add policy for role %d, permission %s: %v", roleId, permCode, err)
			return err
		}
	}

	// Save policies
	err = enforcer.SavePolicy()
	if err != nil {
		logx.Errorf("Failed to save policies: %v", err)
		return err
	}

	return nil
}

// LoadRolePermissions loads role permissions from database
func LoadRolePermissions(ctx context.Context, rolePermModel model.AuthRolePermissionModel, permModel model.AuthPermissionModel, roleId int64) ([]string, error) {
	// Get role permissions
	rolePerms, err := rolePermModel.FindByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}

	if len(rolePerms) == 0 {
		return []string{}, nil
	}

	// Get permission IDs
	permIds := make([]int64, len(rolePerms))
	for i, rp := range rolePerms {
		permIds[i] = rp.PermissionId
	}

	// Get permissions
	permissions, err := permModel.FindByIds(ctx, permIds)
	if err != nil {
		return nil, err
	}

	// Extract permission codes
	permCodes := make([]string, len(permissions))
	for i, p := range permissions {
		permCodes[i] = p.Code
	}

	return permCodes, nil
}
