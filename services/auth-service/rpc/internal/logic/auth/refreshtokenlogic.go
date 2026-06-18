package auth

import (
	"context"
	"fmt"
	"time"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

// RefreshTokenLogic AT 过期无感刷新（docs/specs/auth-design.md §3.4）
//
//  1. 客户端携带 RT 请求 RefreshToken 接口
//  2. 解析 RT 获取 user_id, device_id, jti
//  3. 去 Redis 查询 auth:rt:{user_id}:{device_id}：
//     - 查不到或值不等于 jti → 已注销或已被旋转，返回 403 强制重新登录
//     - 查到且匹配 → 确认 RT 合法
//  4. 重新拉取角色（非破坏性，先于旋转。失败则旧 RT 仍有效）
//  5. RT 旋转（防泄露）——Lua 原子操作：
//     - 删除旧 RT → 写入新 RT（重置 15 天 TTL）
//  6. 生成新 AT（含最新 roles）+ 新 RT → 返回
type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// rotateRTLua RT 旋转原子操作 Lua 脚本
// KEYS[1]: auth:rt:{user_id}:{device_id}
// ARGV[1]: old_jti（预期的旧 jti）
// ARGV[2]: new_jti（新 jti）
// ARGV[3]: ttl（秒）
// 返回: new_jti（成功）或 nil（jti 不匹配，已旋转或已注销）
const rotateRTLua = `
local current = redis.call('GET', KEYS[1])
if current ~= ARGV[1] then
    return nil
end
redis.call('SET', KEYS[1], ARGV[2], 'EX', ARGV[3])
return ARGV[2]
`

// RefreshToken 执行 AT 刷新
func (l *RefreshTokenLogic) RefreshToken(in *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	// 1. 解析 RT，提取信息
	rtString := in.RefreshToken
	rtToken, err := jwt.Parse(rtString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(l.svcCtx.Config.JwtAuth.RefreshSecret), nil
	})
	if err != nil || !rtToken.Valid {
		l.Infof("RefreshToken: invalid RT: %v", err)
		return &authv1.RefreshTokenResponse{
			Base: responsex.NewBaseRespWithError(50003, "RT已失效，请重新登录"),
		}, nil
	}

	rtClaims, ok := rtToken.Claims.(jwt.MapClaims)
	if !ok {
		return &authv1.RefreshTokenResponse{
			Base: responsex.NewBaseRespWithError(50003, "RT已失效，请重新登录"),
		}, nil
	}

	userID := int64(rtClaims["user_id"].(float64))
	deviceID := rtClaims["device_id"].(string)
	oldJti := rtClaims["jti"].(string)

	// 2. 校验 RT 在 Redis 中存在且 jti 匹配
	rtKey := fmt.Sprintf("auth:rt:%d:%s", userID, deviceID)
	currentJti, err := l.svcCtx.RedisClient.Get(l.ctx, rtKey).Result()
	if err != nil || currentJti != oldJti {
		l.Infof("RefreshToken: RT not found or jti mismatch for user=%d, device=%s (current=%s, expected=%s)", userID, deviceID, currentJti, oldJti)
		return &authv1.RefreshTokenResponse{
			Base: responsex.NewBaseRespWithError(50003, "RT已失效，请重新登录"),
		}, nil
	}

	// 3. 重新拉取角色（非破坏性操作，先于旋转。失败则旧 RT 在 Redis 中仍然有效，客户端可重试）
	roles, err := getUserRolesWithCache(l.ctx, l.svcCtx, userID)
	if err != nil {
		l.Errorf("RefreshToken: getUserRolesWithCache failed: %v", err)
		return &authv1.RefreshTokenResponse{
			Base: responsex.NewBaseRespWithError(509504, "获取用户角色失败"),
		}, nil
	}

	// 4. RT 旋转（Lua 原子操作，仅在角色拉取成功后执行）
	now := time.Now()
	newJti := fmt.Sprintf("%d-%d", userID, now.UnixNano())
	refreshExpire := time.Duration(l.svcCtx.Config.JwtAuth.RefreshExpire) * time.Second

	result, err := l.svcCtx.RedisClient.Eval(l.ctx, rotateRTLua,
		[]string{rtKey},                // KEYS[1]
		oldJti,                         // ARGV[1]
		newJti,                         // ARGV[2]
		int(refreshExpire.Seconds()),   // ARGV[3]
	).Result()

	if err != nil || result == nil {
		l.Infof("RefreshToken: rotation failed for user=%d (race condition or already rotated)", userID)
		return &authv1.RefreshTokenResponse{
			Base: responsex.NewBaseRespWithError(50003, "RT已失效，请重新登录"),
		}, nil
	}

	// 5. 生成新 AT（含最新 roles）+ RT
	accessExpire := time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire) * time.Second

	// 新 AT（含 roles）
	atClaims := jwt.MapClaims{
		"user_id": userID,
		"jti":     newJti,
		"roles":   roles,
		"exp":     now.Add(accessExpire).Unix(),
		"iat":     now.Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(l.svcCtx.Config.JwtAuth.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("sign AT failed: %w", err)
	}

	// 新 RT
	rtClaimsNew := jwt.MapClaims{
		"user_id":   userID,
		"device_id": deviceID,
		"jti":       newJti,
		"exp":       now.Add(refreshExpire).Unix(),
		"iat":       now.Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaimsNew)
	newRefreshToken, err := rt.SignedString([]byte(l.svcCtx.Config.JwtAuth.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("sign RT failed: %w", err)
	}

	l.Infof("RefreshToken: rotated RT for user=%d, device=%s, oldJti=%s → newJti=%s, roles=%d",
		userID, deviceID, oldJti, newJti, len(roles))

	return &authv1.RefreshTokenResponse{
		Base:         responsex.NewBaseResp(),
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    now.Add(accessExpire).Unix(),
	}, nil
}
