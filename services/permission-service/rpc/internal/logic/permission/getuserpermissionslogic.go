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
//
//	查询用户活跃 grants（status∈{0,1,2}，含未认证业主发布权限码）→ 收集角色关联的权限 ID → 去重返回
func (l *GetUserPermissionsLogic) GetUserPermissions(in *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
	// 查询用户活跃 grants（T1.5：不再只取 status=2，保证未认证业主的发布权限码在列）
	roles, err := l.svcCtx.UserRoleModel.FindActiveRolesByUserId(l.ctx, in.UserId)
	if err != nil || len(roles) == 0 {
		return &permissionv1.GetUserPermissionsResponse{
			Base:            responsex.NewBaseResp(),
			PermissionCodes: nil,
		}, nil
	}

	// 收集所有角色的权限 ID
	permIdSet := make(map[int64]struct{})
	for _, r := range roles {
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
