package permission

import (
	"context"

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
//   校验角色存在 → 更新字段 → 替换权限列表 → 失效相关用户缓存
func (l *UpdateRoleLogic) UpdateRole(in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
	// 校验角色存在
	existing, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &permissionv1.UpdateRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	// 更新字段
	if in.Name != nil {
		existing.RoleName = *in.Name
	}
	if in.Description != nil {
		existing.Description = sqlNullString(*in.Description)
	}
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
	// 查询所有持有该角色的用户 ID，失效其缓存
	// 由于没有反向查询方法，这里使用 Redis 的 key pattern 删除
	// 实际项目中可以考虑维护 role→users 的映射
	keys, err := l.svcCtx.RedisClient.Keys(l.ctx, "perm:user:*").Result()
	if err != nil {
		l.Errorf("invalidateRoleCache: keys failed: %v", err)
		return
	}
	for _, key := range keys {
		l.svcCtx.RedisClient.Del(l.ctx, key)
	}
	// 同时失效 scope 缓存
	scopeKeys, _ := l.svcCtx.RedisClient.Keys(l.ctx, "perm:scopes:*").Result()
	for _, key := range scopeKeys {
		l.svcCtx.RedisClient.Del(l.ctx, key)
	}
	l.Infof("invalidateRoleCache: invalidated %d perm keys + %d scope keys for role %d",
		len(keys), len(scopeKeys), roleId)
}
