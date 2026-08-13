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
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

// LoginSmsLogic 短信验证码登录（docs/specs/auth-design.md §3.2）
type LoginSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginSmsLogic {
	return &LoginSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// LoginSms 短信验证码登录
//
//  1. 校验短信验证码
//  2. AES 加密手机号 → 查 auth_credential
//  3. 调 User Service GetUserByPhone 获取用户信息
//  4. 校验账号状态（status != 2 禁用）
//  5. 获取已认证角色（Cache-Aside: Redis → 未命中 → gRPC → 回填）
//  6. 签发 AT（含 roles）+ RT → 成功后删除验证码
func (l *LoginSmsLogic) LoginSms(in *authv1.LoginSmsRequest) (*authv1.LoginResponse, error) {
	// 1. 参数校验
	if in.Phone == "" || in.SmsCode == "" {
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50004, "手机号和验证码不能为空"),
		}, nil
	}

	// 2. 从 Redis 校验短信验证码（不立即删除，等登录成功后删除）
	codeKey := fmt.Sprintf("sms:code:%s", in.Phone)
	storedCode, err := l.svcCtx.RedisClient.Get(l.ctx, codeKey).Result()
	if err == redis.Nil {
		l.Infof("LoginSms: code expired for phone=%s", in.Phone[:3]+"****"+in.Phone[len(in.Phone)-4:])
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50004, "验证码已过期，请重新获取"),
		}, nil
	}
	if err != nil {
		l.Errorf("LoginSms: Redis GET %s failed: %v", codeKey, err)
		return nil, err
	}
	if storedCode != in.SmsCode {
		l.Infof("LoginSms: code mismatch for phone=%s", in.Phone[:3]+"****"+in.Phone[len(in.Phone)-4:])
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50004, "验证码错误"),
		}, nil
	}

	// 3. AES 加密手机号，查凭证
	encryptedPhone, err := crypto.AESEncrypt(in.Phone)
	if err != nil {
		l.Errorf("LoginSms: AES encrypt phone failed: %v", err)
		return nil, err
	}

	credential, err := l.svcCtx.CredentialModel.FindByIdentityTypeAndIdentifier(l.ctx, "phone", encryptedPhone)
	if err != nil {
		// 首次短信登录 → 需要先注册。验证码不删除，可用于注册流程
		l.Infof("LoginSms: no credential for phone=%s, need register first", in.Phone[:3]+"****"+in.Phone[len(in.Phone)-4:])
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50001, "该手机号未注册，请先注册"),
		}, nil
	}

	// 4. 获取用户信息
	userResp, err := l.svcCtx.UserServiceRpc.GetUserByPhone(l.ctx, &userv1.GetUserByPhoneRequest{Phone: in.Phone})
	if err != nil || userResp == nil || userResp.Base.GetCode() != 0 {
		l.Errorf("LoginSms: GetUserByPhone rpc failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(509503, "获取用户信息失败"),
		}, nil
	}

	// 4.5 校验账号状态（见 auth-design.md §3.1 步骤 5）
	if userResp.User != nil && userResp.User.Status == 2 {
		l.Infof("LoginSms: user %d is disabled", credential.UserId)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(50005, "账号已被禁用"),
		}, nil
	}

	userId := credential.UserId

	// 5. 获取已认证角色（Cache-Aside: Redis → 未命中 → gRPC → 回填）
	roles, err := getUserRolesWithCache(l.ctx, l.svcCtx, userId)
	if err != nil {
		l.Errorf("LoginSms: getUserRolesWithCache failed: %v", err)
		return &authv1.LoginResponse{
			Base: responsex.NewBaseRespWithError(509504, "获取用户角色失败"),
		}, nil
	}

	// 5.5 端准入判定（签发 Token 前，读 permission GetUserRoles.platforms）
	// SEE: [[is-system-no-permission-shortcut]]
	if err := checkPlatformAccess(l.ctx, l.svcCtx, userId, in.DeviceType); err != nil {
		l.Infof("LoginSms: platform access denied for user=%d, device=%s", userId, in.DeviceType)
		return &authv1.LoginResponse{Base: responsex.NewBaseRespFromError(err)}, nil
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

	rtKey := fmt.Sprintf("auth:rt:%d:%s", userId, in.DeviceId)
	l.svcCtx.RedisClient.Set(l.ctx, rtKey, jti, refreshExpire)
	l.svcCtx.RedisClient.SAdd(l.ctx, fmt.Sprintf("auth:rt:%d:devices", userId), in.DeviceId)

	// 登录成功后删除验证码（防止重复使用）
	if err := l.svcCtx.RedisClient.Del(l.ctx, codeKey).Err(); err != nil {
		l.Errorf("LoginSms: Redis DEL %s failed: %v", codeKey, err)
	}

	l.Infof("LoginSms success: userId=%d, phone=%s", userId, in.Phone[:3]+"****"+in.Phone[len(in.Phone)-4:])

	return &authv1.LoginResponse{
		Base:         responsex.NewBaseResp(),
		AccessToken:  at,
		RefreshToken: rt,
		ExpiresAt:    now.Add(accessExpire).Unix(),
		UserId:       userId,
	}, nil
}
