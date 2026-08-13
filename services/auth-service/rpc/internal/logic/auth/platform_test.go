package auth

import (
	"context"
	"testing"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// classifyDeviceType 端归类
// =============================================================================

func TestClassifyDeviceType(t *testing.T) {
	cases := []struct {
		name       string
		deviceType string
		want       string
	}{
		{"web→pc", "web", "pc"},
		{"admin→pc", "admin", "pc"},
		{"ios→mobile", "ios", "mobile"},
		{"android→mobile", "android", "mobile"},
		{"miniapp→mobile", "miniapp", "mobile"},
		{"空→空", "", ""},
		{"未知→空", "unknown", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, classifyDeviceType(c.deviceType))
		})
	}
}

func TestRoleAllows(t *testing.T) {
	cases := []struct {
		name        string
		platforms   []string
		deviceClass string
		want        bool
	}{
		{"含当前端", []string{"pc", "mobile"}, "pc", true},
		{"不含当前端", []string{"mobile"}, "pc", false},
		{"空 platforms", []string{}, "pc", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, roleAllows(c.platforms, c.deviceClass))
		})
	}
}

// =============================================================================
// checkPlatformAccess 端准入判定
// =============================================================================

func TestCheckPlatformAccess(t *testing.T) {
	newCtx := func(t *testing.T, roles []*permissionv1.UserRoleInfo, err error) *svc.ServiceContext {
		t.Helper()
		perm := &mockPermissionServiceClient{
			GetUserRolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				if err != nil {
					return nil, err
				}
				return &permissionv1.GetUserRolesResponse{Base: okResp(), Roles: roles}, nil
			},
		}
		return &svc.ServiceContext{PermissionClient: perm}
	}

	role := func(platforms ...string) *permissionv1.UserRoleInfo {
		return &permissionv1.UserRoleInfo{Role: &permissionv1.Role{Platforms: platforms}}
	}

	cases := []struct {
		name       string
		deviceType string
		roles      []*permissionv1.UserRoleInfo
		getErr     error
		wantCode   int // 0 = 放行（nil）
	}{
		{"mobile角色+web→拒绝50007", "web", []*permissionv1.UserRoleInfo{role("mobile")}, nil, codePlatformDenied},
		{"mobile角色+android→放行", "android", []*permissionv1.UserRoleInfo{role("mobile")}, nil, 0},
		{"双端角色+web→放行", "web", []*permissionv1.UserRoleInfo{role("pc", "mobile")}, nil, 0},
		{"空platforms角色+web→放行", "web", []*permissionv1.UserRoleInfo{role()}, nil, 0},
		{"未知device_type→放行", "tv", []*permissionv1.UserRoleInfo{role("mobile")}, nil, 0},
		{"零角色→放行", "web", []*permissionv1.UserRoleInfo{}, nil, 0},
		{"多角色其一含当前端→放行", "web", []*permissionv1.UserRoleInfo{role("mobile"), role("pc")}, nil, 0},
		{"GetUserRoles失败→放行", "web", nil, grpc.ErrClientConnClosing, 0},
	}

	// 补充：resp 业务错误（code != 0）→ fail-open 放行（杀 L60 的 ||→&&）
	t.Run("resp业务错误code非0→放行", func(t *testing.T) {
		perm := &mockPermissionServiceClient{
			GetUserRolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: &commonv1.BaseResp{Code: 12345}, Roles: nil}, nil
			},
		}
		svcCtx := &svc.ServiceContext{PermissionClient: perm}
		assert.NoError(t, checkPlatformAccess(context.Background(), svcCtx, 1001, "web"))
	})

	// 补充：info 非 nil 但 Role nil → fail-open（杀 L72 的 ||→&&）
	t.Run("role信息nil→放行", func(t *testing.T) {
		perm := &mockPermissionServiceClient{
			GetUserRolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{
					Base:  okResp(),
					Roles: []*permissionv1.UserRoleInfo{{Role: nil}},
				}, nil
			},
		}
		svcCtx := &svc.ServiceContext{PermissionClient: perm}
		assert.NoError(t, checkPlatformAccess(context.Background(), svcCtx, 1001, "web"))
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := newCtx(t, c.roles, c.getErr)
			err := checkPlatformAccess(context.Background(), ctx, 1001, c.deviceType)
			if c.wantCode == 0 {
				assert.NoError(t, err, "应放行")
				return
			}
			require.Error(t, err)
			ce := errx.FromError(err)
			require.NotNil(t, ce)
			assert.Equal(t, c.wantCode, ce.Code)
		})
	}
}

// mobileOnlyPermRpc 返回「仅 mobile 端」角色的 permission mock（端拒绝用例复用）。
func mobileOnlyPermRpc() *mockPermissionServiceClient {
	return &mockPermissionServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return &permissionv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*permissionv1.UserRoleInfo{{Role: &permissionv1.Role{Platforms: []string{"mobile"}}}},
			}, nil
		},
	}
}
