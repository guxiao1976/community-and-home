package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/model"
	"github.com/guxiao1976/community-auth/rpc/internal/config"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// Mock UserServiceClient
// =============================================================================

type mockUserServiceClient struct {
	CreateUserFn     func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error)
	GetUserByPhoneFn func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error)
	GetUserRolesFn   func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error)
	UpdateUserFn     func(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error)
}

func (m *mockUserServiceClient) CreateUser(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, in, opts...)
	}
	return &userv1.CreateUserResponse{}, nil
}
func (m *mockUserServiceClient) GetUser(ctx context.Context, in *userv1.GetUserRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	return &userv1.GetUserResponse{}, nil
}
func (m *mockUserServiceClient) GetUserByPhone(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
	if m.GetUserByPhoneFn != nil {
		return m.GetUserByPhoneFn(ctx, in, opts...)
	}
	return &userv1.GetUserResponse{}, nil
}
func (m *mockUserServiceClient) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error) {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(ctx, in, opts...)
	}
	return &userv1.UpdateUserResponse{}, nil
}
func (m *mockUserServiceClient) GetUserRoles(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
	if m.GetUserRolesFn != nil {
		return m.GetUserRolesFn(ctx, in, opts...)
	}
	return &userv1.GetUserRolesResponse{}, nil
}

// Remaining methods — auth-service doesn't call these directly
func (m *mockUserServiceClient) BatchUpdateUsers(ctx context.Context, in *userv1.BatchUpdateUsersRequest, opts ...grpc.CallOption) (*userv1.BatchUpdateUsersResponse, error) {
	return &userv1.BatchUpdateUsersResponse{}, nil
}
func (m *mockUserServiceClient) ListUsers(ctx context.Context, in *userv1.ListUsersRequest, opts ...grpc.CallOption) (*userv1.ListUsersResponse, error) {
	return &userv1.ListUsersResponse{}, nil
}
func (m *mockUserServiceClient) GetUsersByIds(ctx context.Context, in *userv1.GetUsersByIdsRequest, opts ...grpc.CallOption) (*userv1.GetUsersByIdsResponse, error) {
	return &userv1.GetUsersByIdsResponse{}, nil
}
func (m *mockUserServiceClient) JoinCommunity(ctx context.Context, in *userv1.JoinCommunityRequest, opts ...grpc.CallOption) (*userv1.JoinCommunityResponse, error) {
	return &userv1.JoinCommunityResponse{}, nil
}
func (m *mockUserServiceClient) LeaveCommunity(ctx context.Context, in *userv1.LeaveCommunityRequest, opts ...grpc.CallOption) (*userv1.LeaveCommunityResponse, error) {
	return &userv1.LeaveCommunityResponse{}, nil
}
func (m *mockUserServiceClient) GetUserMemberships(ctx context.Context, in *userv1.GetUserMembershipsRequest, opts ...grpc.CallOption) (*userv1.GetUserMembershipsResponse, error) {
	return &userv1.GetUserMembershipsResponse{}, nil
}
func (m *mockUserServiceClient) ApplyRole(ctx context.Context, in *userv1.ApplyRoleRequest, opts ...grpc.CallOption) (*userv1.ApplyRoleResponse, error) {
	return &userv1.ApplyRoleResponse{}, nil
}
func (m *mockUserServiceClient) CheckAccess(ctx context.Context, in *userv1.CheckAccessRequest, opts ...grpc.CallOption) (*userv1.CheckAccessResponse, error) {
	return &userv1.CheckAccessResponse{}, nil
}
func (m *mockUserServiceClient) SubmitCertification(ctx context.Context, in *userv1.SubmitCertificationRequest, opts ...grpc.CallOption) (*userv1.SubmitCertificationResponse, error) {
	return &userv1.SubmitCertificationResponse{}, nil
}
func (m *mockUserServiceClient) ReviewCertification(ctx context.Context, in *userv1.ReviewCertificationRequest, opts ...grpc.CallOption) (*userv1.ReviewCertificationResponse, error) {
	return &userv1.ReviewCertificationResponse{}, nil
}
func (m *mockUserServiceClient) ListCertifications(ctx context.Context, in *userv1.ListCertificationsRequest, opts ...grpc.CallOption) (*userv1.ListCertificationsResponse, error) {
	return &userv1.ListCertificationsResponse{}, nil
}
func (m *mockUserServiceClient) GetMyCertifications(ctx context.Context, in *userv1.GetMyCertificationsRequest, opts ...grpc.CallOption) (*userv1.GetMyCertificationsResponse, error) {
	return &userv1.GetMyCertificationsResponse{}, nil
}
func (m *mockUserServiceClient) BindResidence(ctx context.Context, in *userv1.BindResidenceRequest, opts ...grpc.CallOption) (*userv1.BindResidenceResponse, error) {
	return &userv1.BindResidenceResponse{}, nil
}
func (m *mockUserServiceClient) GetResidences(ctx context.Context, in *userv1.GetResidencesRequest, opts ...grpc.CallOption) (*userv1.GetResidencesResponse, error) {
	return &userv1.GetResidencesResponse{}, nil
}
func (m *mockUserServiceClient) UpdateUserModerationStatus(ctx context.Context, in *userv1.UpdateModerationStatusRequest, opts ...grpc.CallOption) (*userv1.UpdateModerationStatusResponse, error) {
	return &userv1.UpdateModerationStatusResponse{}, nil
}
func (m *mockUserServiceClient) GetAppState(ctx context.Context, in *userv1.GetAppStateRequest, opts ...grpc.CallOption) (*userv1.GetAppStateResponse, error) {
	return &userv1.GetAppStateResponse{}, nil
}
func (m *mockUserServiceClient) SetCurrentCommunity(ctx context.Context, in *userv1.SetCurrentCommunityRequest, opts ...grpc.CallOption) (*userv1.SetCurrentCommunityResponse, error) {
	return &userv1.SetCurrentCommunityResponse{}, nil
}

var _ userv1.UserServiceClient = (*mockUserServiceClient)(nil)

// =============================================================================
// Mock PermissionServiceClient
// =============================================================================

type mockPermissionServiceClient struct {
	GetUserRolesFn func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error)
}

func (m *mockPermissionServiceClient) GetUserRoles(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
	if m.GetUserRolesFn != nil {
		return m.GetUserRolesFn(ctx, in, opts...)
	}
	// 默认零角色 → fail-open 放行
	return &permissionv1.GetUserRolesResponse{Base: okResp()}, nil
}

func (m *mockPermissionServiceClient) CheckPermission(ctx context.Context, in *permissionv1.CheckPermissionRequest, opts ...grpc.CallOption) (*permissionv1.CheckPermissionResponse, error) {
	return &permissionv1.CheckPermissionResponse{}, nil
}
func (m *mockPermissionServiceClient) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	return &permissionv1.GetDataScopesResponse{}, nil
}
func (m *mockPermissionServiceClient) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	return &permissionv1.AssertPublishScopeResponse{}, nil
}
func (m *mockPermissionServiceClient) AssignRole(ctx context.Context, in *permissionv1.AssignRoleRequest, opts ...grpc.CallOption) (*permissionv1.AssignRoleResponse, error) {
	return &permissionv1.AssignRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) RevokeRole(ctx context.Context, in *permissionv1.RevokeRoleRequest, opts ...grpc.CallOption) (*permissionv1.RevokeRoleResponse, error) {
	return &permissionv1.RevokeRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) ListRoles(ctx context.Context, in *permissionv1.ListRolesRequest, opts ...grpc.CallOption) (*permissionv1.ListRolesResponse, error) {
	return &permissionv1.ListRolesResponse{}, nil
}
func (m *mockPermissionServiceClient) ListPermissions(ctx context.Context, in *permissionv1.ListPermissionsRequest, opts ...grpc.CallOption) (*permissionv1.ListPermissionsResponse, error) {
	return &permissionv1.ListPermissionsResponse{}, nil
}
func (m *mockPermissionServiceClient) CreateRole(ctx context.Context, in *permissionv1.CreateRoleRequest, opts ...grpc.CallOption) (*permissionv1.CreateRoleResponse, error) {
	return &permissionv1.CreateRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) UpdateRole(ctx context.Context, in *permissionv1.UpdateRoleRequest, opts ...grpc.CallOption) (*permissionv1.UpdateRoleResponse, error) {
	return &permissionv1.UpdateRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) DeleteRole(ctx context.Context, in *permissionv1.DeleteRoleRequest, opts ...grpc.CallOption) (*permissionv1.DeleteRoleResponse, error) {
	return &permissionv1.DeleteRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) GetRole(ctx context.Context, in *permissionv1.GetRoleRequest, opts ...grpc.CallOption) (*permissionv1.GetRoleResponse, error) {
	return &permissionv1.GetRoleResponse{}, nil
}
func (m *mockPermissionServiceClient) GetRolesByIds(ctx context.Context, in *permissionv1.GetRolesByIdsRequest, opts ...grpc.CallOption) (*permissionv1.GetRolesByIdsResponse, error) {
	return &permissionv1.GetRolesByIdsResponse{}, nil
}
func (m *mockPermissionServiceClient) GetUserPermissions(ctx context.Context, in *permissionv1.GetUserPermissionsRequest, opts ...grpc.CallOption) (*permissionv1.GetUserPermissionsResponse, error) {
	return &permissionv1.GetUserPermissionsResponse{}, nil
}
func (m *mockPermissionServiceClient) InvalidateUserCache(ctx context.Context, in *permissionv1.InvalidateUserCacheRequest, opts ...grpc.CallOption) (*permissionv1.InvalidateUserCacheResponse, error) {
	return &permissionv1.InvalidateUserCacheResponse{}, nil
}
func (m *mockPermissionServiceClient) UpdateUserRoleStatus(ctx context.Context, in *permissionv1.UpdateUserRoleStatusRequest, opts ...grpc.CallOption) (*permissionv1.UpdateUserRoleStatusResponse, error) {
	return &permissionv1.UpdateUserRoleStatusResponse{}, nil
}

var _ permissionv1.PermissionServiceClient = (*mockPermissionServiceClient)(nil)

// =============================================================================
// Mock AuthCredentialModel
// =============================================================================

type mockCredentialModel struct {
	InsertFn                          func(ctx context.Context, data *model.AuthCredential) (int64, error)
	FindByIdentityTypeAndIdentifierFn func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error)
}

func (m *mockCredentialModel) Insert(ctx context.Context, data *model.AuthCredential) (int64, error) {
	if m.InsertFn != nil {
		return m.InsertFn(ctx, data)
	}
	return 1, nil
}
func (m *mockCredentialModel) FindOne(ctx context.Context, id int64) (*model.AuthCredential, error) {
	return nil, nil
}
func (m *mockCredentialModel) FindByIdentityTypeAndIdentifier(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
	if m.FindByIdentityTypeAndIdentifierFn != nil {
		return m.FindByIdentityTypeAndIdentifierFn(ctx, identityType, identifier)
	}
	return nil, nil
}
func (m *mockCredentialModel) FindByUserId(ctx context.Context, userId int64) ([]*model.AuthCredential, error) {
	return nil, nil
}
func (m *mockCredentialModel) UpdateCredential(ctx context.Context, id int64, newCredential string) error {
	return nil
}

var _ model.AuthCredentialModel = (*mockCredentialModel)(nil)

// =============================================================================
// Test constants & setup
// =============================================================================

const (
	testAccessSecret  = "test-access-secret-key-for-jwt-signing-32bytes"
	testRefreshSecret = "test-refresh-secret-key-for-jwt-signing-32bytes"
	testAESKey        = "1234567890abcdef1234567890abcdef"
)

func newTestServiceContext(t *testing.T, rds *redis.Client, userRpc userv1.UserServiceClient, credModel model.AuthCredentialModel) *svc.ServiceContext {
	t.Helper()
	return newTestServiceContextWithPermission(t, rds, userRpc, credModel, &mockPermissionServiceClient{})
}

func newTestServiceContextWithPermission(t *testing.T, rds *redis.Client, userRpc userv1.UserServiceClient, credModel model.AuthCredentialModel, permRpc permissionv1.PermissionServiceClient) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		Config: config.Config{
			JwtAuth: config.JwtAuthConfig{
				AccessSecret:  testAccessSecret,
				AccessExpire:  900,
				RefreshSecret: testRefreshSecret,
				RefreshExpire: 1296000,
			},
			AesKey: testAESKey,
		},
		CredentialModel:  credModel,
		RedisClient:      rds,
		UserServiceRpc:   userRpc,
		PermissionClient: permRpc,
	}
}

// testCryptoOnce 缓存 RSA 密钥对/AES 密钥初始化，避免每个测试重复生成 2048-bit RSA
// 密钥（约百毫秒级），显著加速变异测试（gomu 对每个变异体跑一次完整测试集）。
var testCryptoOnce sync.Once

func setupTestCrypto(t *testing.T) {
	t.Helper()
	testCryptoOnce.Do(func() {
		pub, priv, err := crypto.GenerateRSAKeyPair(2048)
		require.NoError(t, err, "GenerateRSAKeyPair")
		require.NoError(t, crypto.InitRSA(pub, priv), "InitRSA")
		require.NoError(t, crypto.InitAES(testAESKey), "InitAES")
	})
}

func rsaEncryptPhone(t *testing.T, phone string) string {
	t.Helper()
	enc, err := crypto.RSAEncrypt(phone)
	require.NoError(t, err, "RSAEncrypt phone")
	return enc
}

func rsaEncryptPassword(t *testing.T, password string) string {
	t.Helper()
	enc, err := crypto.RSAEncrypt(password)
	require.NoError(t, err, "RSAEncrypt password")
	return enc
}

func defaultMockUserRpc(userId int64) *mockUserServiceClient {
	return &mockUserServiceClient{
		CreateUserFn: func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			return &userv1.CreateUserResponse{Base: okResp(), UserId: userId}, nil
		},
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: userId, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{Base: okResp()}, nil
		},
		UpdateUserFn: func(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error) {
			return &userv1.UpdateUserResponse{Base: okResp()}, nil
		},
	}
}

func defaultMockCredentialModel(userId int64) *mockCredentialModel {
	return &mockCredentialModel{
		InsertFn: func(ctx context.Context, data *model.AuthCredential) (int64, error) {
			return 1, nil
		},
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: userId, IdentityType: "phone", Identifier: identifier}, nil
		},
	}
}

func okResp() *commonv1.BaseResp { return &commonv1.BaseResp{Code: 0, Msg: "success"} }
func errResp(code int32, msg string) *commonv1.BaseResp {
	return &commonv1.BaseResp{Code: code, Msg: msg}
}

func parseAT(t *testing.T, at string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(at, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(testAccessSecret), nil
	})
	require.NoError(t, err, "Token 解析应成功")
	require.True(t, token.Valid, "Token 应为有效的 JWT")
	return token.Claims.(jwt.MapClaims)
}
