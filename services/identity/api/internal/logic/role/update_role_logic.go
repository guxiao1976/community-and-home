// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update role
func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleReq) (resp *types.UpdateRoleResp, err error) {
	// Get existing role
	role, err := l.svcCtx.AuthRoleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			logx.Errorf("Role not found: %d", req.Id)
			return nil, err
		}
		logx.Errorf("Failed to get role: %v", err)
		return nil, err
	}

	// Update role fields
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Status != nil {
		role.Status = int64(*req.Status)
	}
	role.UpdatedTime = time.Now()

	// Update role
	err = l.svcCtx.AuthRoleModel.Update(l.ctx, role)
	if err != nil {
		logx.Errorf("Failed to update role: %v", err)
		return nil, err
	}

	// Update permissions if provided
	if req.PermissionIds != nil {
		// Delete existing permissions
		err = l.svcCtx.AuthRolePermissionModel.DeleteByRoleId(l.ctx, req.Id)
		if err != nil {
			logx.Errorf("Failed to delete role permissions: %v", err)
			return nil, err
		}

		// Insert new permissions
		if len(req.PermissionIds) > 0 {
			err = l.svcCtx.AuthRolePermissionModel.BatchInsert(l.ctx, req.Id, req.PermissionIds)
			if err != nil {
				logx.Errorf("Failed to assign permissions to role: %v", err)
				return nil, err
			}
		}
	}

	return &types.UpdateRoleResp{
		Success: true,
	}, nil
}

