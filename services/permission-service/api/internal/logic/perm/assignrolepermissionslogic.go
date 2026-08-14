package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermissionsLogic {
	return &AssignRolePermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AssignRolePermissions 更新角色的权限列表（替换模式：先删后插）
//
// REQ-PLAT-8 一致性不变量：先 GetRole 读取角色当前 platforms，UpdateRole 时显式保留，
// 防止 D3「platforms 无条件覆盖」语义下权限独立分配清空角色端限制（安全回归）。
// SEE: [[verify-api-before-calling]] — 先 GetRole 验证角色存在，再 UpdateRole
// SEE: [[rpc-callback-must-check-response-base]] — GetRole/UpdateRole 响应均须 base-check
func (l *AssignRolePermissionsLogic) AssignRolePermissions(req *types.AssignRolePermissionsReq, roleId int64) error {
	// 先读角色当前 platforms
	getResp, err := l.svcCtx.PermissionRpc.GetRole(l.ctx, &permissionv1.GetRoleRequest{Id: roleId})
	if err != nil {
		return err
	}
	// GetRole 业务错误（如 60001 角色不存在）→ abort，不调用 UpdateRole（防带空 platforms 清空端限制）
	if err := responsex.ToError(getResp.Base); err != nil {
		return err
	}

	var platforms []string
	if getResp.Role != nil {
		platforms = getResp.Role.Platforms
	}

	updateResp, err := l.svcCtx.PermissionRpc.UpdateRole(l.ctx, &permissionv1.UpdateRoleRequest{
		Id:            roleId,
		PermissionIds: req.PermissionIds,
		Platforms:     platforms, // 显式保留现有 platforms
	})
	if err != nil {
		return err
	}
	// UpdateRole 响应 base-check（REQ-UPDATE-3）
	return responsex.ToError(updateResp.Base)
}
