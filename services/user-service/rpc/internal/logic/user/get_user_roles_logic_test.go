package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permissionmocks "github.com/guxiao1976/api-proto/gen/go/permission/v1/mocks"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/redis/redistest"
)

// getRolesTestSvc 创建带 permission mock 的 ServiceContext（复用 certTestSvc）
func getRolesTestSvc(t *testing.T) (*svc.ServiceContext, *permissionmocks.MockPermissionServiceClient) {
	return certTestSvc(t)
}

func TestGetUserRoles_ReturnsCertified(t *testing.T) {
	// U-G-01: 返回用户已认证角色（status=2）
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	req := &userv1.GetUserRolesRequest{
		UserId:     7001,
		VerfStatus: int32Ptr(2), // 只取已认证
	}
	resp, err := logic.GetUserRoles(req)

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "owner", resp.Roles[0].RoleCode)
	assert.Equal(t, int32(2), resp.Roles[0].VerfStatus)
}

func TestGetUserRoles_AllStatuses(t *testing.T) {
	// U-G-02: 不传 verf_status 返回所有状态
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
			{Role: &permissionv1.Role{Id: 2, Code: "property_admin"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)
	assert.Len(t, resp.Roles, 2)
}

func TestGetUserRoles_NoRoles(t *testing.T) {
	// U-G-03: 无角色返回空列表
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 9999})

	require.NoError(t, err)
	assert.Empty(t, resp.Roles)
}

func TestGetUserRoles_ApprovedFiltersNonApproved(t *testing.T) {
	// U-G-04: 只查已认证(status=2)时，非已认证角色应被过滤
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
			{Role: &permissionv1.Role{Id: 2, Code: "property_admin"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(2)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "property_admin", resp.Roles[0].RoleCode)
}

func TestGetUserRoles_Pending_NoFilter(t *testing.T) {
	// U-G-05: VerfStatus=1（待审核）→ 不按已认证过滤，全部返回
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
			{Role: &permissionv1.Role{Id: 2, Code: "property_admin"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(1)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 2)
}

func TestGetUserRoles_Rejected_NoFilter(t *testing.T) {
	// U-G-06: VerfStatus=3（已驳回）→ 非已认证，不按已认证过滤
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(3)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
}

func TestGetUserRoles_GetUserRolesError(t *testing.T) {
	// U-G-07: permission GetUserRoles 失败 → 透传错误
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission service down")).Times(1)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestGetUserRoles_VerifiedAtPopulated(t *testing.T) {
	// U-G-08: VerifiedAt>0 时应回填到 proto
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 2, VerifiedAt: 1700000000},
			{Role: &permissionv1.Role{Id: 2, Code: "tenant"}, ScopeType: "community", ScopeId: 2001, Status: 0},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 2)
	assert.Equal(t, int64(1700000000), resp.Roles[0].VerifiedAt, "VerifiedAt>0 应回填")
	assert.Equal(t, int64(0), resp.Roles[1].VerifiedAt, "VerifiedAt=0 不应回填")
}

func TestGetUserRoles_PermissionClientNil(t *testing.T) {
	// U-G-09: PermissionClient 为 nil → 50000 系统繁忙
	svc := testSvc(t) // 不注入 permission mock

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)
	assert.Equal(t, int32(50000), resp.Base.Code)
}

func TestGetUserRoles_ApprovedFilterAndExpiresAt(t *testing.T) {
	// U-G-10: 已认证查询中 status!=2 的角色应被过滤；ExpiresAt 透传
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 3, ExpiresAt: 9999},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(2)})

	require.NoError(t, err)
	assert.Empty(t, resp.Roles, "status=3 角色在已认证查询中应被过滤")
}

func TestGetUserRoles_CacheHit_ReturnsCached(t *testing.T) {
	// U-G-11: 已认证查询命中缓存 → 直接返回缓存，不调 permission
	svc, permMock := getRolesTestSvc(t)
	svc.Redis = redistest.CreateRedis(t)

	cached := &userv1.GetUserRolesResponse{
		Base: responsex.NewBaseResp(),
		Roles: []*userv1.MembershipRole{
			{UserId: 7001, RoleCode: "cached_owner", CommunityId: 2001, VerfStatus: 2},
		},
	}
	setRolesToCache(context.Background(), svc.Redis, 7001, cached, 300)

	// 命中缓存 → GetUserRoles 不应被调用
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Times(0)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(2)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "cached_owner", resp.Roles[0].RoleCode)
}

func TestGetUserRoles_NonApproved_IgnoresCache(t *testing.T) {
	// U-G-12: 非已认证查询不读缓存 → 走 permission
	svc, permMock := getRolesTestSvc(t)
	svc.Redis = redistest.CreateRedis(t)

	// 预置缓存（应被忽略）
	cached := &userv1.GetUserRolesResponse{
		Base:  responsex.NewBaseResp(),
		Roles: []*userv1.MembershipRole{{UserId: 7001, RoleCode: "cached_owner", VerfStatus: 2}},
	}
	setRolesToCache(context.Background(), svc.Redis, 7001, cached, 300)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "fresh_owner", 0, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "fresh_owner", resp.Roles[0].RoleCode, "非已认证应走 permission 而非缓存")
}

func TestGetUserRoles_Approved_WritesCache(t *testing.T) {
	// U-G-13: 已认证查询回填缓存
	svc, permMock := getRolesTestSvc(t)
	svc.Redis = redistest.CreateRedis(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(2)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)

	// 缓存应已写入
	cached := getRolesFromCache(context.Background(), svc.Redis, 7001)
	require.NotNil(t, cached, "已认证查询应回填缓存")
	require.Len(t, cached.Roles, 1)
	assert.Equal(t, "owner", cached.Roles[0].RoleCode)
}

func TestGetUserRoles_NonApproved_NoCacheWrite(t *testing.T) {
	// U-G-14: 非已认证查询不写缓存
	svc, permMock := getRolesTestSvc(t)
	svc.Redis = redistest.CreateRedis(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 0, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	_, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)

	cached := getRolesFromCache(context.Background(), svc.Redis, 7001)
	assert.Nil(t, cached, "非已认证查询不应写缓存")
}

func TestGetUserRoles_SysConfigTTLApplied(t *testing.T) {
	// U-G-15: SysConfig 配置的缓存 TTL 生效（默认 300 → 配置 600）
	mr := miniredis.RunT(t)
	// 预置 sys_config Hash
	mr.HSet("sys_config", "user.cache.roles_ttl_seconds", `{"value":"600","type":"number"}`)

	svc, permMock := getRolesTestSvc(t)
	svc.Redis = redis.New(mr.Addr())
	svc.SysConfig = sysconfig.MustInit(redis.RedisConf{Host: mr.Addr(), Type: "node"}, "sys_config", nil)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001, VerfStatus: int32Ptr(2)})

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)

	ttl := mr.TTL("auth:roles:7001")
	assert.Equal(t, 600*time.Second, ttl, "应使用 sysconfig 配置的 TTL=600")
}
