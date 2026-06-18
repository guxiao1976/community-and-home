package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/model"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// §2.2 账密登录 (Login) 测试
// =============================================================================

func TestLogin_Success(t *testing.T) {
	// A-L-01: 正常登录（有 owner 角色）
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)

	userId := int64(3001)
	encPwd := rsaEncryptPassword(t, "CorrectPassword1")
	encPhone := rsaEncryptPhone(t, "13900139000")
	hashedPwd, _ := crypto.HashPassword("CorrectPassword1")

	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: userId, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{
					{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2},
				},
			}, nil
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: userId, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)

	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: encPhone, EncryptedPassword: encPwd, DeviceId: "web_001", DeviceType: "web",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, userId, resp.UserId)

	// 验证 AT 含 roles
	claims := parseAT(t, resp.AccessToken)
	rolesJSON, _ := json.Marshal(claims["roles"])
	var roles []roleEntry
	json.Unmarshal(rolesJSON, &roles)
	require.Len(t, roles, 1)
	assert.Equal(t, "owner", roles[0].R)
	assert.Equal(t, int64(1001), roles[0].C)

	// RT 持久化
	rtKey := fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")
	assert.True(t, mr.Exists(rtKey))
}

func TestLogin_MultiRoles(t *testing.T) {
	// A-L-02: 多角色登录
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.FlushAll() // 确保干净状态

	userId := int64(3002)
	encPwd := rsaEncryptPassword(t, "Password123")
	encPhone := rsaEncryptPhone(t, "13900139001")
	hashedPwd, _ := crypto.HashPassword("Password123")

	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: userId, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{
					{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2},
					{RoleCode: "committee", CommunityId: 1001, VerfStatus: 2},
				},
			}, nil
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: userId, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: encPhone, EncryptedPassword: encPwd, DeviceId: "web_001",
	})

	require.NoError(t, err)
	claims := parseAT(t, resp.AccessToken)
	rolesJSON, _ := json.Marshal(claims["roles"])
	var roles []roleEntry
	json.Unmarshal(rolesJSON, &roles)
	assert.Len(t, roles, 2)
}

func TestLogin_PhoneNotRegistered(t *testing.T) {
	// A-L-03: 凭证不存在
	_, rdb := setupRedis(t)
	setupTestCrypto(t)

	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3003), credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: rsaEncryptPhone(t, "13900000000"), EncryptedPassword: rsaEncryptPassword(t, "test"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50001), resp.Base.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	// A-L-04: 密码错误
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	hashedCorrect, _ := crypto.HashPassword("CorrectPassword")

	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: 3004, Credential: hashedCorrect}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3004), credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: rsaEncryptPhone(t, "13900139000"), EncryptedPassword: rsaEncryptPassword(t, "WrongPassword"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50001), resp.Base.Code)
}

func TestLogin_RSADecryptPhoneFailed(t *testing.T) {
	// A-L-05: RSA 解密 phone 失败
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3005), defaultMockCredentialModel(3005))
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: "invalid_rsa_encrypted", EncryptedPassword: rsaEncryptPassword(t, "test"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50001), resp.Base.Code)
}

func TestLogin_GetUserRolesFailed(t *testing.T) {
	// A-L-08: GetUserRoles gRPC 失败
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	hashedPwd, _ := crypto.HashPassword("Test123")

	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: 3008, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return nil, fmt.Errorf("gRPC connection error")
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: 3008, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: rsaEncryptPhone(t, "13900139000"), EncryptedPassword: rsaEncryptPassword(t, "Test123"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(509504), resp.Base.Code)
}

func TestLogin_GetUserByPhoneFailed(t *testing.T) {
	// A-L-09: GetUserByPhone gRPC 失败
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	hashedPwd, _ := crypto.HashPassword("Test123")

	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return nil, fmt.Errorf("gRPC timeout")
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: 3009, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: rsaEncryptPhone(t, "13900139000"), EncryptedPassword: rsaEncryptPassword(t, "Test123"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(509503), resp.Base.Code)
}

func TestLogin_EmptyCredentials(t *testing.T) {
	// A-L-10/A-L-11: 空手机号/密码
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3010), defaultMockCredentialModel(3010))
	logic := NewLoginLogic(context.Background(), svcCtx)

	resp, err := logic.Login(&authv1.LoginRequest{EncryptedPhone: "", EncryptedPassword: "test"})
	require.NoError(t, err)
	assert.Equal(t, int32(50400), resp.Base.Code)

	resp, err = logic.Login(&authv1.LoginRequest{EncryptedPhone: "test", EncryptedPassword: ""})
	require.NoError(t, err)
	assert.Equal(t, int32(50400), resp.Base.Code)
}

func TestLogin_MerchantRole_CZero(t *testing.T) {
	// A-L-14: merchant 角色 c=0
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	hashedPwd, _ := crypto.HashPassword("Merchant123")

	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: 3014, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "merchant", CommunityId: 0, VerfStatus: 2}},
			}, nil
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: 3014, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&authv1.LoginRequest{
		EncryptedPhone: rsaEncryptPhone(t, "13900139002"), EncryptedPassword: rsaEncryptPassword(t, "Merchant123"), DeviceId: "web_001",
	})

	require.NoError(t, err)
	claims := parseAT(t, resp.AccessToken)
	rolesJSON, _ := json.Marshal(claims["roles"])
	var roles []roleEntry
	json.Unmarshal(rolesJSON, &roles)
	require.Len(t, roles, 1)
	assert.Equal(t, "merchant", roles[0].R)
	assert.Equal(t, int64(0), roles[0].C)
}

// =============================================================================
// §2.3 短信验证码登录 (LoginSms) 测试
// =============================================================================

func TestLoginSms_Success(t *testing.T) {
	// A-S-01: 正常短信登录
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13900139000", "123456")

	userId := int64(3101)
	hashedPwd, _ := crypto.HashPassword("")
	userRpc := &mockUserServiceClient{
		GetUserByPhoneFn: func(ctx context.Context, in *userv1.GetUserByPhoneRequest, opts ...grpc.CallOption) (*userv1.GetUserResponse, error) {
			return &userv1.GetUserResponse{Base: okResp(), User: &userv1.User{Id: userId, Status: 1}}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2}},
			}, nil
		},
	}
	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return &model.AuthCredential{Id: 1, UserId: userId, Credential: hashedPwd}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	logic := NewLoginSmsLogic(context.Background(), svcCtx)
	resp, err := logic.LoginSms(&authv1.LoginSmsRequest{
		Phone: "13900139000", SmsCode: "123456", DeviceId: "web_001",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.False(t, mr.Exists("sms:code:13900139000"), "验证码使用后应删除")
}

func TestLoginSms_CodeExpired(t *testing.T) {
	// A-S-03: 验证码过期
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3103), defaultMockCredentialModel(3103))
	logic := NewLoginSmsLogic(context.Background(), svcCtx)
	resp, err := logic.LoginSms(&authv1.LoginSmsRequest{
		Phone: "13900139000", SmsCode: "expired_code",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50004), resp.Base.Code)
}

func TestLoginSms_CodeMismatch(t *testing.T) {
	// A-S-02: 验证码不匹配，不被删除
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13900139000", "123456")
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3102), defaultMockCredentialModel(3102))
	logic := NewLoginSmsLogic(context.Background(), svcCtx)
	resp, err := logic.LoginSms(&authv1.LoginSmsRequest{
		Phone: "13900139000", SmsCode: "wrong_code",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50004), resp.Base.Code)
	assert.True(t, mr.Exists("sms:code:13900139000"))
}

func TestLoginSms_PhoneNotRegistered(t *testing.T) {
	// A-S-04: 验证码正确但手机号未注册
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13900139000", "123456")

	credModel := &mockCredentialModel{
		FindByIdentityTypeAndIdentifierFn: func(ctx context.Context, identityType, identifier string) (*model.AuthCredential, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3104), credModel)
	logic := NewLoginSmsLogic(context.Background(), svcCtx)
	resp, err := logic.LoginSms(&authv1.LoginSmsRequest{
		Phone: "13900139000", SmsCode: "123456",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(50001), resp.Base.Code)
	assert.True(t, mr.Exists("sms:code:13900139000"), "凭证不存在时验证码应保留（Q-04 修复）")
}

// =============================================================================
// §2.4 Token 刷新 (RefreshToken) 测试
// =============================================================================

func TestRefreshToken_Success(t *testing.T) {
	// A-T-01: 正常刷新
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3201)

	oldJti := fmt.Sprintf("%d-%d", userId, time.Now().UnixNano())
	rtKey := fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")
	mr.Set(rtKey, oldJti)

	now := time.Now()
	rtClaims := jwt.MapClaims{
		"user_id": float64(userId), "device_id": "web_001", "jti": oldJti,
		"exp": float64(now.Add(15 * 24 * time.Hour).Unix()), "iat": float64(now.Unix()),
	}
	oldRT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString([]byte(testRefreshSecret))

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2}},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))
	logic := NewRefreshTokenLogic(context.Background(), svcCtx)
	resp, err := logic.RefreshToken(&authv1.RefreshTokenRequest{RefreshToken: oldRT})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.NotEqual(t, oldRT, resp.RefreshToken, "新 RT 应与旧 RT 不同")
	assert.True(t, mr.Exists(rtKey), "新 RT 应写入 Redis")

	claims := parseAT(t, resp.AccessToken)
	rolesJSON, _ := json.Marshal(claims["roles"])
	var roles []roleEntry
	json.Unmarshal(rolesJSON, &roles)
	assert.Len(t, roles, 1)
}

func TestRefreshToken_JTIMismatch(t *testing.T) {
	// A-T-03: RT jti 不匹配（已被旋转）
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3203)

	rtKey := fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")
	mr.Set(rtKey, "different_jti_value")

	now := time.Now()
	rtClaims := jwt.MapClaims{
		"user_id": float64(userId), "device_id": "web_001",
		"jti": fmt.Sprintf("%d-original", userId),
		"exp": float64(now.Add(15 * 24 * time.Hour).Unix()), "iat": float64(now.Unix()),
	}
	originalRT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString([]byte(testRefreshSecret))

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(userId), defaultMockCredentialModel(userId))
	logic := NewRefreshTokenLogic(context.Background(), svcCtx)
	resp, err := logic.RefreshToken(&authv1.RefreshTokenRequest{RefreshToken: originalRT})

	require.NoError(t, err)
	assert.Equal(t, int32(50003), resp.Base.Code)
}

func TestRefreshToken_RolesUpdated(t *testing.T) {
	// A-T-06: 刷新后角色更新（角色变更后刷新）
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3206)

	oldJti := fmt.Sprintf("%d-%d", userId, time.Now().UnixNano())
	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001"), oldJti)

	now := time.Now()
	rtClaims := jwt.MapClaims{
		"user_id": float64(userId), "device_id": "web_001", "jti": oldJti,
		"exp": float64(now.Add(15 * 24 * time.Hour).Unix()), "iat": float64(now.Unix()),
	}
	oldRT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString([]byte(testRefreshSecret))

	// 模拟角色变化：之前只有 owner，现在新增 committee
	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{
					{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2},
					{RoleCode: "committee", CommunityId: 1001, VerfStatus: 2},
				},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))
	logic := NewRefreshTokenLogic(context.Background(), svcCtx)
	resp, err := logic.RefreshToken(&authv1.RefreshTokenRequest{RefreshToken: oldRT})

	require.NoError(t, err)
	claims := parseAT(t, resp.AccessToken)
	rolesJSON, _ := json.Marshal(claims["roles"])
	var roles []roleEntry
	json.Unmarshal(rolesJSON, &roles)
	assert.Len(t, roles, 2, "刷新后 AT 应包含最新角色")
}

func TestRefreshToken_GetUserRolesFailed(t *testing.T) {
	// A-T-09: 刷新时 GetUserRoles 失败
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3209)

	oldJti := fmt.Sprintf("%d-%d", userId, time.Now().UnixNano())
	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001"), oldJti)

	now := time.Now()
	rtClaims := jwt.MapClaims{
		"user_id": float64(userId), "device_id": "web_001", "jti": oldJti,
		"exp": float64(now.Add(15 * 24 * time.Hour).Unix()), "iat": float64(now.Unix()),
	}
	oldRT, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString([]byte(testRefreshSecret))

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return nil, fmt.Errorf("gRPC error")
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))
	logic := NewRefreshTokenLogic(context.Background(), svcCtx)
	resp, err := logic.RefreshToken(&authv1.RefreshTokenRequest{RefreshToken: oldRT})

	require.NoError(t, err)
	assert.Equal(t, int32(509504), resp.Base.Code)

	// Q-02 修复验证：角色拉取失败时，旧 RT 仍存在（旋转发生在角色拉取成功之后）
	rtKey2 := fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")
	assert.True(t, mr.Exists(rtKey2), "角色拉取失败时旧 RT 应保留（Q-02 修复）")
}

// =============================================================================
// §2.5 注销 (Logout) 测试
// =============================================================================

func TestLogout_Success(t *testing.T) {
	// A-O-01: 正常注销
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3301)
	now := time.Now()
	jti := fmt.Sprintf("%d-%d", userId, now.UnixNano())

	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001"), jti)

	atClaims := jwt.MapClaims{
		"user_id": float64(userId), "jti": jti,
		"exp": float64(now.Add(900 * time.Second).Unix()), "iat": float64(now.Unix()),
	}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(testAccessSecret))

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(userId), defaultMockCredentialModel(userId))
	logic := NewLogoutLogic(context.Background(), svcCtx)
	resp, err := logic.Logout(&authv1.LogoutRequest{
		AccessToken: at, UserId: userId, DeviceId: "web_001",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.True(t, mr.Exists(fmt.Sprintf("auth:at:blacklist:%s", jti)), "AT 应在黑名单中")
	assert.False(t, mr.Exists(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")), "RT 应删除")
}

func TestLogout_KickAllDevices(t *testing.T) {
	// A-O-04: 强踢全设备
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(3304)
	now := time.Now()
	jti := fmt.Sprintf("%d-%d", userId, now.UnixNano())

	atClaims := jwt.MapClaims{
		"user_id": float64(userId), "jti": jti,
		"exp": float64(now.Add(900 * time.Second).Unix()), "iat": float64(now.Unix()),
	}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(testAccessSecret))

	devicesKey := fmt.Sprintf("auth:rt:%d:devices", userId)
	mr.SAdd(devicesKey, "web_001", "ios_001", "android_001")
	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001"), "jti-web")
	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "ios_001"), "jti-ios")
	mr.Set(fmt.Sprintf("auth:rt:%d:%s", userId, "android_001"), "jti-android")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(userId), defaultMockCredentialModel(userId))
	logic := NewLogoutLogic(context.Background(), svcCtx)
	resp, err := logic.Logout(&authv1.LogoutRequest{
		AccessToken: at, UserId: userId, DeviceId: "web_001", KickAllDevices: true,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.False(t, mr.Exists(fmt.Sprintf("auth:rt:%d:%s", userId, "web_001")))
	assert.False(t, mr.Exists(fmt.Sprintf("auth:rt:%d:%s", userId, "ios_001")))
	assert.False(t, mr.Exists(fmt.Sprintf("auth:rt:%d:%s", userId, "android_001")))
}

// =============================================================================
// §2.6 Token 验证 (ValidateToken) 测试
// =============================================================================

func TestValidateToken_Valid(t *testing.T) {
	// A-V-01: 有效 AT
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	now := time.Now()
	userId := int64(3401)

	atClaims := jwt.MapClaims{
		"user_id": float64(userId), "jti": "valid_jti",
		"exp": float64(now.Add(900 * time.Second).Unix()), "iat": float64(now.Unix()),
	}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(testAccessSecret))

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(userId), defaultMockCredentialModel(userId))
	logic := NewValidateTokenLogic(context.Background(), svcCtx)
	resp, err := logic.ValidateToken(&authv1.ValidateTokenRequest{AccessToken: at})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, userId, resp.UserId)
}

func TestValidateToken_Expired(t *testing.T) {
	// A-V-02: AT 已过期
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	now := time.Now()

	atClaims := jwt.MapClaims{
		"user_id": float64(3402), "jti": "expired_jti",
		"exp": float64(now.Add(-1 * time.Hour).Unix()), "iat": float64(now.Add(-2 * time.Hour).Unix()),
	}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(testAccessSecret))

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3402), defaultMockCredentialModel(3402))
	logic := NewValidateTokenLogic(context.Background(), svcCtx)
	resp, _ := logic.ValidateToken(&authv1.ValidateTokenRequest{AccessToken: at})
	assert.False(t, resp.Valid)
}

func TestValidateToken_Blacklisted(t *testing.T) {
	// A-V-03: AT 在黑名单中
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	now := time.Now()
	jti := "blacklisted_jti"
	mr.Set(fmt.Sprintf("auth:at:blacklist:%s", jti), "1")

	atClaims := jwt.MapClaims{
		"user_id": float64(3403), "jti": jti,
		"exp": float64(now.Add(900 * time.Second).Unix()), "iat": float64(now.Unix()),
	}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(testAccessSecret))

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3403), defaultMockCredentialModel(3403))
	logic := NewValidateTokenLogic(context.Background(), svcCtx)
	resp, err := logic.ValidateToken(&authv1.ValidateTokenRequest{AccessToken: at})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	// A-V-04: AT 签名无效
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3404), defaultMockCredentialModel(3404))
	logic := NewValidateTokenLogic(context.Background(), svcCtx)
	resp, err := logic.ValidateToken(&authv1.ValidateTokenRequest{AccessToken: "invalid.jwt.token"})
	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	// A-V-05: 空 AT
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(3405), defaultMockCredentialModel(3405))
	logic := NewValidateTokenLogic(context.Background(), svcCtx)
	resp, _ := logic.ValidateToken(&authv1.ValidateTokenRequest{AccessToken: ""})
	assert.False(t, resp.Valid)
}
