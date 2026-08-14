package permission

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// UpdateRole 更新角色
//
//	校验角色存在 → 系统角色 status 门禁（先于字段与 platforms 校验）→ 校验 platforms → 更新字段 → 替换权限列表 → 失效相关用户缓存
//
// 系统角色字段级策略（REQ-UPDATE-4 / D1）：
//   - name/description/platforms/sort_order/permission_ids 可编辑（is_system=1 放行）
//   - status 仍拦截（60004 原子拒绝，先于任何字段应用、先于 platforms 校验；无部分写入）
//   - 60004 message 收窄为「系统角色状态不可修改」（仅本路径；DeleteRole 路径 60004 语义不改）
//
// SEE: [[is-system-no-permission-shortcut]] — 字段级放行不改变权限模型（无特权短路）
func (l *UpdateRoleLogic) UpdateRole(in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
	// 校验角色存在
	existing, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &permissionv1.UpdateRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	// 系统角色 status 门禁（校验顺序钉死：先于任何字段应用、先于 platforms 校验，原子拒绝）
	if existing.IsSystem == 1 && in.Status != nil {
		return &permissionv1.UpdateRoleResponse{
			Base: responsex.NewBaseRespWithError(60004, "系统角色状态不可修改"),
		}, nil
	}

	// 校验允许登录端（REQ-PLAT-4）：非法值 → 60008 原子拒绝，任何字段不落库
	// SEE: [[error-code-literal-bypasses-qa-gate]] — 60008 用命名常量
	platforms, err := validatePlatforms(in.Platforms)
	if err != nil {
		return &permissionv1.UpdateRoleResponse{
			Base: responsex.NewBaseRespFromError(err),
		}, nil
	}

	// 更新字段
	if in.Name != nil {
		existing.RoleName = *in.Name
	}
	if in.Description != nil {
		existing.Description = sqlNullString(*in.Description)
	}
	// platforms 无条件覆盖（D3）：空列表 = 显式清空（fail-open）
	existing.Platforms = joinPlatforms(platforms)
	// status 仅非系统角色应用（系统角色 status 已在上方门禁拦截）
	if in.Status != nil {
		existing.Status = int64(*in.Status)
	}
	if in.SortOrder != nil {
		existing.SortOrder = int64(*in.SortOrder)
	}

	if err := l.svcCtx.RoleModel.Update(l.ctx, existing); err != nil {
		l.Errorf("UpdateRole: update failed: %v", err)
		return nil, err
	}

	// 替换权限列表（仅当传入了权限 ID 时才操作）
	if len(in.PermissionIds) > 0 {
		if err := l.svcCtx.RolePermissionModel.DeleteByRoleId(l.ctx, in.Id); err != nil {
			l.Errorf("UpdateRole: clear permissions failed: %v", err)
			return nil, err
		}
		var records []*model.RelRolePermission
		for _, pid := range in.PermissionIds {
			records = append(records, &model.RelRolePermission{
				RoleId:       in.Id,
				PermissionId: pid,
			})
		}
		if err := l.svcCtx.RolePermissionModel.BatchInsert(l.ctx, records); err != nil {
			l.Errorf("UpdateRole: batch insert permissions failed: %v", err)
			return nil, err
		}
	}

	// 失效所有持有该角色的用户缓存
	l.invalidateRoleCache(in.Id)

	l.Infof("UpdateRole success: id=%d", in.Id)

	// 查询最新权限
	rps, _ := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, in.Id)
	var permIds []int64
	for _, rp := range rps {
		permIds = append(permIds, rp.PermissionId)
	}

	return &permissionv1.UpdateRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: roleToPbWithPermissions(l.ctx, existing, permIds, l.svcCtx.PermissionModel),
	}, nil
}

// invalidateRoleCache 失效所有持有该角色的用户权限缓存
func (l *UpdateRoleLogic) invalidateRoleCache(roleId int64) {
	// 查询所有持有该角色的用户 ID
	userRoles, err := l.svcCtx.UserRoleModel.FindByRoleId(l.ctx, roleId)
	if err != nil {
		l.Errorf("invalidateRoleCache: find user roles failed: %v", err)
		return
	}

	// 失效每个用户的权限缓存
	deletedCount := 0
	userSet := make(map[int64]bool)
	for _, ur := range userRoles {
		if userSet[ur.UserId] {
			continue // 同一用户只删除一次
		}
		userSet[ur.UserId] = true

		// 删除该用户的权限缓存和数据范围缓存
		keys := []string{
			fmt.Sprintf("perm:user:%d", ur.UserId),
			fmt.Sprintf("perm:scopes:%d:community", ur.UserId),
			fmt.Sprintf("perm:scopes:%d:building", ur.UserId),
			fmt.Sprintf("perm:scopes:%d:unit", ur.UserId),
			fmt.Sprintf("perm:scopes:%d:grid", ur.UserId),
		}
		if deleted, err := l.svcCtx.RedisClient.Del(l.ctx, keys...).Result(); err == nil {
			deletedCount += int(deleted)
		}
	}

	l.Infof("invalidateRoleCache: invalidated %d users (%d keys) for role %d",
		len(userSet), deletedCount, roleId)
}
