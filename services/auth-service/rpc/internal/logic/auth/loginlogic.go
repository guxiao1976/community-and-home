package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/core/logx"
)

// LoginLogic 账密登录（docs/specs/auth-design.md §3.1）
//
//  1. RSA 私钥解密 phone + password
//  2. AES 加密 phone → 查 auth_credential
//  3. bcrypt 校验 password
//  4. 调用 user-service.GetUserByPhone 获取用户信息
//  5. 校验 status != 2（禁用）
//  6. 获取已认证角色（Cache-Aside: Redis → gRPC → 回填）
//  7. 签发 AT（含 roles）+ RT
//  8. Redis 持久化 RT + 加入设备集合
//  9. 返回 AT, RT, expires_at, user_id
type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 执行账密登录
func (l *LoginLogic) Login(in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	// 1. 参数校验
	if in.EncryptedPhone == "" || in.EncryptedPassword == "" {
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50400, "手机号和密码不能为空"),
		}, nil
	}

	// 2. RSA 解密客户端传来的密文
	phone, err := crypto.RSADecrypt(in.EncryptedPhone)
	if err != nil {
		l.Infof("Login: RSA decrypt phone failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50001, "登录失败，请检查账号密码"),
		}, nil
	}
	password, err := crypto.RSADecrypt(in.EncryptedPassword)
	if err != nil {
		l.Infof("Login: RSA decrypt password failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50001, "登录失败，请检查账号密码"),
		}, nil
	}

	// 3. AES 加密手机号 → 作为 identifier 查询凭证
	encryptedPhone, err := crypto.AESEncrypt(phone)
	if err != nil {
		l.Errorf("Login: AES encrypt phone failed: %v", err)
		return nil, err
	}

	credential, err := l.svcCtx.CredentialModel.FindByIdentityTypeAndIdentifier(l.ctx, "phone", encryptedPhone)
	if err != nil {
		l.Infof("Login: credential not found for phone=%s", phone[:3]+"****"+phone[len(phone)-4:])
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50001, "手机号或密码错误"),
		}, nil
	}

	// 4. bcrypt 校验密码
	if !crypto.CheckPassword(password, credential.Credential) {
		l.Infof("Login: password mismatch for userId=%d", credential.UserId)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50001, "手机号或密码错误"),
		}, nil
	}

	// 5. 调用 User Service 获取用户信息
	userResp, err := l.svcCtx.UserServiceRpc.GetUserByPhone(l.ctx, &userv1.GetUserByPhoneRequest{
		Phone: phone,
	})
	if err != nil || userResp == nil || userResp.Base.GetCode() != 0 {
		l.Errorf("Login: GetUserByPhone rpc failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(509503, "获取用户信息失败"),
		}, nil
	}

	// 5.5 校验账号状态（见 auth-design.md §3.1 步骤 5）
	if userResp.User != nil && userResp.User.Status == 2 {
		l.Infof("Login: user %d is disabled", credential.UserId)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50005, "账号已被禁用"),
		}, nil
	}

	// 6. 获取已认证角色（Cache-Aside: Redis → 未命中 → gRPC → 回填）
	roles, err := getUserRolesWithCache(l.ctx, l.svcCtx, credential.UserId)
	if err != nil {
		l.Errorf("Login: getUserRolesWithCache failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(509504, "获取用户角色失败"),
		}, nil
	}

	// 6.5 端准入判定（签发 Token 前，读 permission GetUserRoles.platforms）
	// SEE: [[is-system-no-permission-shortcut]]
	if err := checkPlatformAccess(l.ctx, l.svcCtx, credential.UserId, in.DeviceType); err != nil {
		l.Infof("Login: platform access denied for user=%d, device=%s", credential.UserId, in.DeviceType)
		return &authv1.LoginResponse{Base: responsex.NewBaseRespFromError(err)}, nil
	}

	// 7. 签发 AT + RT（AT 携带 roles）
	now := time.Now()
	accessExpire := time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire) * time.Second
	refreshExpire := time.Duration(l.svcCtx.Config.JwtAuth.RefreshExpire) * time.Second
	jti := fmt.Sprintf("%d-%d", credential.UserId, now.UnixNano())

	// AT（15 分钟，含 roles）
	atClaims := jwt.MapClaims{
		"user_id": credential.UserId,
		"jti":     jti,
		"roles":   roles,
		"exp":     now.Add(accessExpire).Unix(),
		"iat":     now.Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(l.svcCtx.Config.JwtAuth.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("sign AT failed: %w", err)
	}

	// RT（15 天）
	rtClaims := jwt.MapClaims{
		"user_id":   credential.UserId,
		"device_id": in.DeviceId,
		"jti":       jti,
		"exp":       now.Add(refreshExpire).Unix(),
		"iat":       now.Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(l.svcCtx.Config.JwtAuth.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("sign RT failed: %w", err)
	}

	// 8. Redis 持久化 RT
	rtKey := fmt.Sprintf("auth:rt:%d:%s", credential.UserId, in.DeviceId)
	if err := l.svcCtx.RedisClient.Set(l.ctx, rtKey, jti, refreshExpire).Err(); err != nil {
		l.Errorf("Login: Redis SET %s failed: %v", rtKey, err)
		return nil, fmt.Errorf("服务内部错误")
	}

	// 加入设备集合（替代 KEYS 命令）
	devicesKey := fmt.Sprintf("auth:rt:%d:devices", credential.UserId)
	l.svcCtx.RedisClient.SAdd(l.ctx, devicesKey, in.DeviceId)

	l.Infof("Login success: userId=%d, device=%s/%s, jti=%s",
		credential.UserId, in.DeviceType, in.DeviceId, jti)

	return &authv1.LoginResponse{
		Base:         responsex.NewBaseResp(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(accessExpire).Unix(),
		UserId:       credential.UserId,
	}, nil
}
