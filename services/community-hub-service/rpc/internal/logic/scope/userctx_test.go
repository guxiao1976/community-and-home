package scope

import (
	"context"
	"errors"
	"testing"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Task 1.8 测试：PublishRolesFrom（多角色优先序 + 无发布角色空集 + 传输错误 fail-closed）+ PublishRoleToString。

func TestPublishRolesFrom(t *testing.T) {
	verified := func(roleCode string, status int32) *permissionv1.UserRoleInfo {
		ur := verifiedGrant(roleCode, ScopeTypeCommunity, 1001, status)
		return ur
	}

	tests := []struct {
		name     string
		rolesFn  func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error)
		wantErr  bool
		want     []string
	}{
		{
			name: "多角色按优先序 grid_worker > community_admin > committee > property_admin",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verified(RolePropertyAdmin, UserRoleStatusVerified),
					verified(RoleGridWorker, UserRoleStatusVerified),
					verified(RoleCommittee, UserRoleStatusVerified),
				}}, nil
			},
			want: []string{RoleGridWorker, RoleCommittee, RolePropertyAdmin},
		},
		{
			name: "无发布角色（owner/tenant）→ 空集",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verified("owner", UserRoleStatusVerified),
					verified("tenant", UserRoleStatusVerified),
				}}, nil
			},
			want: nil,
		},
		{
			name: "level-2 未认证 grant 不计入",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verified(RoleCommittee, 1), // 待审
					verified(RoleGridWorker, 2),
				}}, nil
			},
			want: []string{RoleGridWorker},
		},
		{
			name: "过期 grant 不计入",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				expired := verified(RoleGridWorker, UserRoleStatusVerified)
				expired.ExpiresAt = time.Now().Unix() - 3600
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{expired}}, nil
			},
			want: nil,
		},
		{
			name: "同角色多 scope grant 去重",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verified(RoleGridWorker, UserRoleStatusVerified),
					verified(RoleGridWorker, UserRoleStatusVerified),
				}}, nil
			},
			want: []string{RoleGridWorker},
		},
		{
			name: "GetUserRoles 传输错误 fail-closed",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return nil, errors.New("permission unavailable")
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perm := &fakeRolesPerm{rolesFn: tc.rolesFn}
			got, err := PublishRolesFrom(context.Background(), perm, 100)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPublishRoleToString(t *testing.T) {
	tests := []struct {
		roleCode string
		want     string
	}{
		{RoleGridWorker, "grid_officer"},
		{RoleCommunityAdmin, "community"},
		{RoleCommittee, "committee"},
		{RolePropertyAdmin, "property"},
		{"owner", ""},
		{"unknown", ""},
	}
	for _, tc := range tests {
		t.Run(tc.roleCode, func(t *testing.T) {
			assert.Equal(t, tc.want, PublishRoleToString(tc.roleCode))
		})
	}
}
