// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package role

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete role
func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(req *types.DeleteRoleReq) (resp *types.DeleteRoleResp, err error) {
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

	// Prevent deletion of system roles
	if role.IsSystem == 1 {
		logx.Errorf("Cannot delete system role: %d", req.Id)
		return nil, errors.New("cannot delete system role")
	}

	// Soft delete: set delete_time
	role.DeleteTime = sql.NullTime{Time: time.Now(), Valid: true}
	err = l.svcCtx.AuthRoleModel.Update(l.ctx, role)
	if err != nil {
		logx.Errorf("Failed to delete role: %v", err)
		return nil, err
	}

	// Delete role permissions
	err = l.svcCtx.AuthRolePermissionModel.DeleteByRoleId(l.ctx, req.Id)
	if err != nil {
		logx.Errorf("Failed to delete role permissions: %v", err)
		return nil, err
	}

	return &types.DeleteRoleResp{
		Success: true,
	}, nil
}

