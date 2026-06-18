package auth

import (
	"context"
	"fmt"
	"testing"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// §2.1 注册 (Register) 测试
// =============================================================================

func setupRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return mr, rdb
}

func TestRegister_Success(t *testing.T) {
	// A-R-01: 正常注册（带密码）
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	userId := int64(2001)
	userRpc := defaultMockUserRpc(userId)
	credModel := defaultMockCredentialModel(userId)
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)

	encPwd := rsaEncryptPassword(t, "TestPass123")
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", EncryptedPassword: encPwd,
		Nickname: "张三", DeviceId: "web_chrome_001", DeviceType: "web",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.NotZero(t, resp.UserId)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Greater(t, resp.ExpiresAt, int64(0))

	// 验证码使用后应被删除
	assert.False(t, mr.Exists("sms:code:13800138000"))

	// RT 写入 Redis
	rtKey := fmt.Sprintf("auth:rt:%d:%s", resp.UserId, "web_chrome_001")
	assert.True(t, mr.Exists(rtKey))

	// 设备加入设备集合
	devicesKey := fmt.Sprintf("auth:rt:%d:devices", resp.UserId)
	isMember, _ := rdb.SIsMember(context.Background(), devicesKey, "web_chrome_001").Result()
	assert.True(t, isMember)

	// AT 的 roles 为空数组
	claims := parseAT(t, resp.AccessToken)
	assert.Equal(t, float64(resp.UserId), claims["user_id"])
	assert.NotEmpty(t, claims["jti"])
	roles, ok := claims["roles"].([]interface{})
	assert.True(t, ok, "roles 应为数组")
	assert.Empty(t, roles, "新用户 roles 应为空")
}

func TestRegister_NoPassword(t *testing.T) {
	// A-R-02: 正常注册（无密码）
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138001", "654321")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2002), defaultMockCredentialModel(2002))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138001", SmsCode: "654321", DeviceId: "ios_001",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestRegister_CodeExpired(t *testing.T) {
	// A-R-03: 验证码已过期
	_, rdb := setupRedis(t)
	setupTestCrypto(t)

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2003), defaultMockCredentialModel(2003))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "expired_code",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(50004), resp.Base.Code)
	assert.Contains(t, resp.Base.Msg, "过期")
}

func TestRegister_CodeMismatch(t *testing.T) {
	// A-R-04: 验证码错误
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2004), defaultMockCredentialModel(2004))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "wrong_code",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(50004), resp.Base.Code)
	assert.Contains(t, resp.Base.Msg, "错误")
	assert.True(t, mr.Exists("sms:code:13800138000"), "错误验证码不应删除")
}

func TestRegister_PhoneAlreadyRegistered(t *testing.T) {
	// A-R-05: 手机号已注册
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	userRpc := &mockUserServiceClient{
		CreateUserFn: func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			return &userv1.CreateUserResponse{Base: errResp(509001, "手机号已注册")}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(2005))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(509001), resp.Base.Code)
}

func TestRegister_RSADecryptPasswordFailed_Saga(t *testing.T) {
	// A-R-06: RSA 密码解密失败 → Saga 补偿
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	compensateCalled := false
	userRpc := &mockUserServiceClient{
		CreateUserFn: func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			return &userv1.CreateUserResponse{Base: okResp(), UserId: 2006}, nil
		},
		UpdateUserFn: func(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error) {
			if in.Status != nil && *in.Status == 3 {
				compensateCalled = true
			}
			return &userv1.UpdateUserResponse{}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(2006))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456",
		EncryptedPassword: "invalid_base64_rsa!!!",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(50400), resp.Base.Code)
	assert.True(t, compensateCalled, "Saga 补偿应被调用")
}

func TestRegister_CredentialInsertFailed_Saga(t *testing.T) {
	// A-R-07: Credential 写入失败 → Saga
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	compensateCalled := false
	userRpc := &mockUserServiceClient{
		CreateUserFn: func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			return &userv1.CreateUserResponse{Base: okResp(), UserId: 2007}, nil
		},
		UpdateUserFn: func(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error) {
			if in.Status != nil && *in.Status == 3 {
				compensateCalled = true
			}
			return &userv1.UpdateUserResponse{}, nil
		},
	}
	credModel := &mockCredentialModel{
		InsertFn: func(ctx context.Context, data *model.AuthCredential) (int64, error) {
			return 0, fmt.Errorf("db insert error")
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, credModel)
	encPwd := rsaEncryptPassword(t, "TestPass123")

	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", EncryptedPassword: encPwd,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(509001), resp.Base.Code)
	assert.True(t, compensateCalled)
}

func TestRegister_EmptyPhone(t *testing.T) {
	// A-R-08: 空手机号
	_, rdb := setupRedis(t)
	setupTestCrypto(t)

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2008), defaultMockCredentialModel(2008))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{Phone: "", SmsCode: "123456"})

	require.NoError(t, err)
	assert.Equal(t, int32(50004), resp.Base.Code)
}

func TestRegister_NewUserRolesEmpty(t *testing.T) {
	// A-R-09: 新用户 AT roles=[]
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2009), defaultMockCredentialModel(2009))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", DeviceId: "test_device",
	})

	require.NoError(t, err)
	claims := parseAT(t, resp.AccessToken)
	roles := claims["roles"]
	assert.NotNil(t, roles)
	rolesArr, ok := roles.([]interface{})
	assert.True(t, ok)
	assert.Empty(t, rolesArr)
}

func TestRegister_GetUserRolesFailed_NonFatal(t *testing.T) {
	// A-R-10: GetUserRoles 失败不阻塞注册
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	userRpc := &mockUserServiceClient{
		CreateUserFn: func(ctx context.Context, in *userv1.CreateUserRequest, opts ...grpc.CallOption) (*userv1.CreateUserResponse, error) {
			return &userv1.CreateUserResponse{Base: okResp(), UserId: 2010}, nil
		},
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return nil, fmt.Errorf("gRPC timeout")
		},
		UpdateUserFn: func(ctx context.Context, in *userv1.UpdateUserRequest, opts ...grpc.CallOption) (*userv1.UpdateUserResponse, error) {
			return &userv1.UpdateUserResponse{}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(2010))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", DeviceId: "test_device",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "GetUserRoles 失败不应阻塞注册")
	assert.NotEmpty(t, resp.AccessToken)
}

func TestRegister_TokenKeyIsolation(t *testing.T) {
	// A-R-11: 双 Token 签名密钥隔离
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2011), defaultMockCredentialModel(2011))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", DeviceId: "test_device",
	})

	require.NoError(t, err)

	// AT 用 AccessSecret 验证通过
	parseAT(t, resp.AccessToken)

	// RT 用 RefreshSecret 验证通过
	rtToken, err := jwtParse(resp.RefreshToken, testRefreshSecret)
	require.NoError(t, err)
	assert.True(t, rtToken.Valid)

	// RT 不能用 AccessSecret 验证
	_, err = jwtParse(resp.RefreshToken, testAccessSecret)
	assert.Error(t, err)
}

func TestRegister_JTIAndDeviceFormat(t *testing.T) {
	// A-R-12: 验证 jti 格式和 RT 中的 device_id
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.Set("sms:code:13800138000", "123456")

	svcCtx := newTestServiceContext(t, rdb, defaultMockUserRpc(2012), defaultMockCredentialModel(2012))
	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&authv1.RegisterRequest{
		Phone: "13800138000", SmsCode: "123456", DeviceId: "android_device_xyz",
	})

	require.NoError(t, err)

	claims := parseAT(t, resp.AccessToken)
	jti := claims["jti"].(string)
	assert.Contains(t, jti, fmt.Sprintf("%d", resp.UserId), "jti 应含 user_id")
	assert.Contains(t, jti, "-", "jti 格式应为 {user_id}-{unix_nano}")

	rtClaims := parseRT(t, resp.RefreshToken)
	assert.Equal(t, "android_device_xyz", rtClaims["device_id"])
}

// jwtParse parses a JWT with the given secret.
func jwtParse(tokenStr, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// parseRT parses an RT and returns its claims.
func parseRT(t *testing.T, rt string) jwt.MapClaims {
	t.Helper()
	token, err := jwtParse(rt, testRefreshSecret)
	require.NoError(t, err, "RT 解析应成功")
	require.True(t, token.Valid, "RT 应为有效的 JWT")
	return token.Claims.(jwt.MapClaims)
}
