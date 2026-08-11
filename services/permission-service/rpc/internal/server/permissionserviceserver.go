package server

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/rpc/internal/logic/permission"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
)

// PermissionServiceServer 权限中心 gRPC Server
type PermissionServiceServer struct {
	svcCtx *svc.ServiceContext
	permissionv1.UnimplementedPermissionServiceServer
}

func NewPermissionServiceServer(svcCtx *svc.ServiceContext) *PermissionServiceServer {
	return &PermissionServiceServer{svcCtx: svcCtx}
}

// CheckPermission 鉴权检查（spec/permission.md 逻辑流 3）
func (s *PermissionServiceServer) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest) (*permissionv1.CheckPermissionResponse, error) {
	l := permission.NewCheckPermissionLogic(ctx, s.svcCtx)
	return l.CheckPermission(in)
}

// GetDataScopes 获取数据范围（spec/permission.md 逻辑流 2）
func (s *PermissionServiceServer) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest) (*permissionv1.GetDataScopesResponse, error) {
	l := permission.NewGetDataScopesLogic(ctx, s.svcCtx)
	return l.GetDataScopes(in)
}

// AssignRole 分配角色（spec/permission.md 逻辑流 1）
func (s *PermissionServiceServer) AssignRole(ctx context.Context, in *permissionv1.AssignRoleRequest) (*permissionv1.AssignRoleResponse, error) {
	l := permission.NewAssignRoleLogic(ctx, s.svcCtx)
	return l.AssignRole(in)
}

// RevokeRole 撤销角色
func (s *PermissionServiceServer) RevokeRole(ctx context.Context, in *permissionv1.RevokeRoleRequest) (*permissionv1.RevokeRoleResponse, error) {
	l := permission.NewRevokeRoleLogic(ctx, s.svcCtx)
	return l.RevokeRole(in)
}

// GetUserRoles 查询用户角色
func (s *PermissionServiceServer) GetUserRoles(ctx context.Context, in *permissionv1.GetUserRolesRequest) (*permissionv1.GetUserRolesResponse, error) {
	l := permission.NewGetUserRolesLogic(ctx, s.svcCtx)
	return l.GetUserRoles(in)
}

// ListRoles 角色列表
func (s *PermissionServiceServer) ListRoles(ctx context.Context, in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
	l := permission.NewListRolesLogic(ctx, s.svcCtx)
	return l.ListRoles(in)
}

// ListPermissions 权限树
func (s *PermissionServiceServer) ListPermissions(ctx context.Context, in *permissionv1.ListPermissionsRequest) (*permissionv1.ListPermissionsResponse, error) {
	l := permission.NewListPermissionsLogic(ctx, s.svcCtx)
	return l.ListPermissions(in)
}

// CreateRole 创建角色
func (s *PermissionServiceServer) CreateRole(ctx context.Context, in *permissionv1.CreateRoleRequest) (*permissionv1.CreateRoleResponse, error) {
	l := permission.NewCreateRoleLogic(ctx, s.svcCtx)
	return l.CreateRole(in)
}

// UpdateRole 更新角色
func (s *PermissionServiceServer) UpdateRole(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
	l := permission.NewUpdateRoleLogic(ctx, s.svcCtx)
	return l.UpdateRole(in)
}

// DeleteRole 删除角色
func (s *PermissionServiceServer) DeleteRole(ctx context.Context, in *permissionv1.DeleteRoleRequest) (*permissionv1.DeleteRoleResponse, error) {
	l := permission.NewDeleteRoleLogic(ctx, s.svcCtx)
	return l.DeleteRole(in)
}

// GetRole 获取角色详情
func (s *PermissionServiceServer) GetRole(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
	l := permission.NewGetRoleLogic(ctx, s.svcCtx)
	return l.GetRole(in)
}

// GetRolesByIds 批量查询角色
func (s *PermissionServiceServer) GetRolesByIds(ctx context.Context, in *permissionv1.GetRolesByIdsRequest) (*permissionv1.GetRolesByIdsResponse, error) {
	l := permission.NewGetRolesByIdsLogic(ctx, s.svcCtx)
	return l.GetRolesByIds(in)
}

// GetUserPermissions 获取用户权限编码列表
func (s *PermissionServiceServer) GetUserPermissions(ctx context.Context, in *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
	l := permission.NewGetUserPermissionsLogic(ctx, s.svcCtx)
	return l.GetUserPermissions(in)
}

// InvalidateUserCache 失效用户权限缓存
func (s *PermissionServiceServer) InvalidateUserCache(ctx context.Context, in *permissionv1.InvalidateUserCacheRequest) (*permissionv1.InvalidateUserCacheResponse, error) {
	l := permission.NewInvalidateUserCacheLogic(ctx, s.svcCtx)
	return l.InvalidateUserCache(in)
}

// UpdateUserRoleStatus 更新用户角色的生命周期状态
func (s *PermissionServiceServer) UpdateUserRoleStatus(ctx context.Context, in *permissionv1.UpdateUserRoleStatusRequest) (*permissionv1.UpdateUserRoleStatusResponse, error) {
	l := permission.NewUpdateUserRoleStatusLogic(ctx, s.svcCtx)
	return l.UpdateUserRoleStatus(in)
}
