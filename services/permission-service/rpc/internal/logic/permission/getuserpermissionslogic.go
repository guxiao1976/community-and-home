package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetUserPermissions 获取用户的所有权限编码列表
//   查询用户活跃角色 → 收集角色关联的权限 ID → 查询权限编码 → 去重返回
func (l *GetUserPermissionsLogic) GetUserPermissions(in *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
	// 查询用户活跃角色
	roles, err := l.svcCtx.UserRoleModel.FindActiveByUserId(l.ctx, in.UserId)
	if err != nil || len(roles) == 0 {
		return &permissionv1.GetUserPermissionsResponse{
			Base:            responsex.NewBaseResp(),
			PermissionCodes: nil,
		}, nil
	}

	// 收集所有角色的权限 ID
	permIdSet := make(map[int64]struct{})
	for _, r := range roles {
		// 系统角色 → 拥有所有权限
		if r.IsSystem == 1 {
			allPerms, err := l.svcCtx.PermissionModel.FindAll(l.ctx)
			if err == nil {
				for _, p := range allPerms {
					permIdSet[p.Id] = struct{}{}
				}
			}
			break // 系统角色不需要再查其他角色
		}

		rps, _ := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, r.RoleId)
		for _, rp := range rps {
			permIdSet[rp.PermissionId] = struct{}{}
		}
	}

	if len(permIdSet) == 0 {
		return &permissionv1.GetUserPermissionsResponse{
			Base:            responsex.NewBaseResp(),
			PermissionCodes: nil,
		}, nil
	}

	// 转换为 ID 列表
	var permIds []int64
	for pid := range permIdSet {
		permIds = append(permIds, pid)
	}

	// 查询权限编码
	perms, err := l.svcCtx.PermissionModel.FindByIds(l.ctx, permIds)
	if err != nil {
		return nil, err
	}

	// 提取编码并去重
	var codes []string
	for _, p := range perms {
		codes = append(codes, p.Code)
	}

	return &permissionv1.GetUserPermissionsResponse{
		Base:            responsex.NewBaseResp(),
		PermissionCodes: codes,
	}, nil
}
