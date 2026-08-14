package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// CreateRole 创建角色
//
//	校验角色编码唯一性 → 校验 platforms → 插入 sys_role → 批量插入 rel_role_permission
func (l *CreateRoleLogic) CreateRole(in *permissionv1.CreateRoleRequest) (*permissionv1.CreateRoleResponse, error) {
	// 校验编码唯一性
	existing, err := l.svcCtx.RoleModel.FindByCode(l.ctx, in.Code)
	if err == nil && existing != nil {
		return &permissionv1.CreateRoleResponse{
			Base: responsex.NewBaseRespWithError(60006, "角色编码已存在"),
		}, nil
	}

	// 校验允许登录端（REQ-PLAT-4）：非法值 → 60008 原子拒绝，不 Insert
	// SEE: [[error-code-literal-bypasses-qa-gate]] — 60008 用命名常量
	platforms, err := validatePlatforms(in.Platforms)
	if err != nil {
		return &permissionv1.CreateRoleResponse{
			Base: responsex.NewBaseRespFromError(err),
		}, nil
	}

	// 插入角色
	role := &model.SysRole{
		RoleCode:    in.Code,
		RoleName:    in.Name,
		Description: sqlNullString(in.Description),
		Status:      1, // 默认启用
		Platforms:   joinPlatforms(platforms),
	}
	if in.SortOrder > 0 {
		role.SortOrder = int64(in.SortOrder)
	}

	roleId, err := l.svcCtx.RoleModel.Insert(l.ctx, role)
	if err != nil {
		l.Errorf("CreateRole: insert role failed: %v", err)
		return nil, err
	}

	// 批量插入权限关联
	if len(in.PermissionIds) > 0 {
		if err := l.svcCtx.RolePermissionModel.DeleteByRoleId(l.ctx, roleId); err != nil {
			l.Errorf("CreateRole: clear old permissions failed: %v", err)
		}
		var records []*model.RelRolePermission
		for _, pid := range in.PermissionIds {
			records = append(records, &model.RelRolePermission{
				RoleId:       roleId,
				PermissionId: pid,
			})
		}
		if err := l.svcCtx.RolePermissionModel.BatchInsert(l.ctx, records); err != nil {
			l.Errorf("CreateRole: batch insert permissions failed: %v", err)
			return nil, err
		}
	}

	// 查询创建后的完整角色
	created, err := l.svcCtx.RoleModel.FindOne(l.ctx, roleId)
	if err != nil {
		return nil, err
	}

	l.Infof("CreateRole success: id=%d, code=%s", roleId, in.Code)

	return &permissionv1.CreateRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: roleToPbWithPermissions(l.ctx, created, in.PermissionIds, l.svcCtx.PermissionModel),
	}, nil
}
