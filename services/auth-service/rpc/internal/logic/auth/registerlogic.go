package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/model"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// RegisterLogic 注册（docs/specs/auth-design.md §3.3）
//
//  1. 校验短信验证码
//  2. 调用 User Service CreateUser → 获取 user_id
//  3. AES 加密手机号 → 写入 auth_credential（identifier = AES(phone), credential = bcrypt(password)）
//  4. Saga 补偿：credential 写入失败 → 调用 User Service 删除用户
//  5. 获取已认证角色（Cache-Aside，新用户 roles=[]）
//  6. 签发 AT（含 roles）+ RT → 返回
type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Register 执行注册
func (l *RegisterLogic) Register(in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	// 1. 校验短信验证码
	if in.Phone == "" || in.SmsCode == "" {
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(50004, "手机号和验证码不能为空"),
		}, nil
	}
	// 从 Redis 读取验证码
	codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
	storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
	if err != nil {
		if err == redis.Nil {
			return &authv1.RegisterResponse{
				Base: responsex.NewBaseRespWithError(50004, "验证码已过期，请重新获取"),
			}, nil
		}
		l.Errorf("Register: Redis GET %s failed: %v", codeKey, err)
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(50004, "系统繁忙，请稍后再试"),
		}, nil
	}
	if storedCode != in.SmsCode {
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(50004, "验证码错误"),
		}, nil
	}
	// 验证通过，删除验证码（防重放）
	if err := l.svcCtx.RedisClient.Del(l.ctx, codeKey).Err(); err != nil {
		l.Errorf("Register: Redis DEL %s failed: %v", codeKey, err)
	}

	// 2. 调用 User Service 创建用户档案
	createUserResp, err := l.svcCtx.UserServiceRpc.CreateUser(l.ctx, &userv1.CreateUserRequest{
		Phone:    in.Phone,
		Nickname: in.Nickname,
		UserType: 1, // 默认居民
	})
	if err != nil {
		l.Errorf("Register: User.CreateUser rpc call failed: %v", err)
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(509001, "注册失败，请稍后重试"),
		}, nil
	}
	if createUserResp == nil || createUserResp.Base.GetCode() != 0 {
		code := int32(509001)
		msg := "注册失败，请稍后重试"
		if createUserResp != nil && createUserResp.Base != nil {
			code = createUserResp.Base.GetCode()
			msg = createUserResp.Base.GetMsg()
		}
		l.Errorf("Register: User.CreateUser business error: code=%d, msg=%s", code, msg)
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(code, msg),
		}, nil
	}
	userId := createUserResp.UserId

	// 3. RSA 解密密码 → bcrypt 哈希
	var bcryptHash string
	if in.EncryptedPassword != "" {
		password, err := crypto.RSADecrypt(in.EncryptedPassword)
		if err != nil {
			// 用户已创建，补偿删除
			l.compensateUser(userId)
			return &authv1.RegisterResponse{
				Base: responsex.NewBaseRespWithError(50400, "密码格式错误"),
			}, nil
		}
		bcryptHash, err = crypto.HashPassword(password)
		if err != nil {
			l.compensateUser(userId)
			return nil, fmt.Errorf("hash password failed: %w", err)
		}
	}

	// 4. AES 加密手机号 → 写入 auth_credential
	encryptedPhone, err := crypto.AESEncrypt(in.Phone)
	if err != nil {
		l.compensateUser(userId)
		return nil, fmt.Errorf("AES encrypt phone failed: %w", err)
	}

	_, err = l.svcCtx.CredentialModel.Insert(l.ctx, &model.AuthCredential{
		UserId:       userId,
		IdentityType: "phone",
		Identifier:   encryptedPhone, // AES 加密的手机号
		Credential:   bcryptHash,
	})
	if err != nil {
		// Saga 补偿：删除已创建的用户
		l.Errorf("Register: insert credential failed: %v, compensating...", err)
		l.compensateUser(userId)
		return &authv1.RegisterResponse{
			Base: responsex.NewBaseRespWithError(509001, "注册失败，请稍后重试"),
		}, nil
	}

	// 5. 获取已认证角色（Cache-Aside。新注册用户通常无角色，roles=[]）
	roles, err := getUserRolesWithCache(l.ctx, l.svcCtx, userId)
	if err != nil {
		// 获取角色失败不阻塞注册，签发空角色 Token
		l.Infof("Register: getUserRolesWithCache failed (non-fatal): %v", err)
		roles = []roleEntry{}
	}

	// 5.5 端准入判定（签发 Token 前，读 permission GetUserRoles.platforms）
	// SEE: [[is-system-no-permission-shortcut]]
	if err := checkPlatformAccess(l.ctx, l.svcCtx, userId, in.DeviceType); err != nil {
		l.Infof("Register: platform access denied for user=%d, device=%s", userId, in.DeviceType)
		return &authv1.RegisterResponse{Base: responsex.NewBaseRespFromError(err)}, nil
	}

	// 6. 签发 AT（含 roles）+ RT
	now := time.Now()
	accessExpire := time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire) * time.Second
	refreshExpire := time.Duration(l.svcCtx.Config.JwtAuth.RefreshExpire) * time.Second
	jti := fmt.Sprintf("%d-%d", userId, now.UnixNano())

	atClaims := jwt.MapClaims{"user_id": userId, "jti": jti, "roles": roles, "exp": now.Add(accessExpire).Unix(), "iat": now.Unix()}
	at, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString([]byte(l.svcCtx.Config.JwtAuth.AccessSecret))

	rtClaims := jwt.MapClaims{"user_id": userId, "device_id": in.DeviceId, "jti": jti, "exp": now.Add(refreshExpire).Unix(), "iat": now.Unix()}
	rt, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims).SignedString([]byte(l.svcCtx.Config.JwtAuth.RefreshSecret))

	// 7. 存储 RT
	rtKey := fmt.Sprintf("auth:rt:%d:%s", userId, in.DeviceId)
	l.svcCtx.RedisClient.Set(l.ctx, rtKey, jti, refreshExpire)
	l.svcCtx.RedisClient.SAdd(l.ctx, fmt.Sprintf("auth:rt:%d:devices", userId), in.DeviceId)

	l.Infof("Register success: userId=%d, phone=%s", userId, in.Phone[:3]+"****"+in.Phone[len(in.Phone)-4:])

	return &authv1.RegisterResponse{
		Base:         responsex.NewBaseResp(),
		UserId:       userId,
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresAt:    now.Add(accessExpire).Unix(),
	}, nil
}

// compensateUser Saga 补偿：将用户状态设为已删除
func (l *RegisterLogic) compensateUser(userId int64) {
	status := int32(3) // deleted
	_, err := l.svcCtx.UserServiceRpc.UpdateUser(l.ctx, &userv1.UpdateUserRequest{
		Id:     userId,
		Status: &status,
	})
	if err != nil {
		l.Errorf("Register: saga compensate failed for userId=%d: %v", userId, err)
	}
}
