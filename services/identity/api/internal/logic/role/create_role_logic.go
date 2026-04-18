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

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create role
func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleReq) (resp *types.CreateRoleResp, err error) {
	// Check if role code already exists
	existingRole, err := l.svcCtx.AuthRoleModel.FindOneByCode(l.ctx, req.Code)
	if err != nil && err != model.ErrNotFound {
		logx.Errorf("Failed to check role code: %v", err)
		return nil, err
	}
	if existingRole != nil {
		logx.Errorf("Role code already exists: %s", req.Code)
		return nil, err
	}

	// Create role
	now := time.Now()
	role := &model.AuthRole{
		Name:        req.Name,
		Code:        req.Code,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		IsSystem:    0, // Custom role
		SortOrder:   0,
		Status:      1, // Active
		CreatedBy:   0, // TODO: Get from JWT context
		CreatedTime: now,
		UpdatedTime: now,
	}

	result, err := l.svcCtx.AuthRoleModel.Insert(l.ctx, role)
	if err != nil {
		logx.Errorf("Failed to create role: %v", err)
		return nil, err
	}

	roleId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("Failed to get role id: %v", err)
		return nil, err
	}

	// Assign permissions to role
	if len(req.PermissionIds) > 0 {
		err = l.svcCtx.AuthRolePermissionModel.BatchInsert(l.ctx, roleId, req.PermissionIds)
		if err != nil {
			logx.Errorf("Failed to assign permissions to role: %v", err)
			return nil, err
		}
	}

	return &types.CreateRoleResp{
		Id: roleId,
	}, nil
}

