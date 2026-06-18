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

// LogoutLogic 主动注销（docs/specs/auth-design.md §3.5）
//
//  1. 解析当前 AT 提取 jti 和剩余过期时间 remaining_ttl
//  2. 拉黑当前 AT: SET auth:at:blacklist:{jti} 1 EX {remaining_ttl}
//     写入失败 → 返回错误，不允许静默忽略（安全关键路径）
//  3. 清除当前设备 RT: DEL auth:rt:{user_id}:{device_id}
//     RT 删除失败不阻塞注销（AT 已拉黑），仅记录日志
//  4. 如果 kick_all_devices=true: 从设备集 SMEMBERS 获取所有设备 → 逐个 DEL RT
type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Logout 执行注销
func (l *LogoutLogic) Logout(in *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	// 1. 解析 AT 提取 jti 和过期时间（不验证 exp，只提取 payload）
	atToken, _, err := new(jwt.Parser).ParseUnverified(in.AccessToken, jwt.MapClaims{})
	if err != nil {
		return &authv1.LogoutResponse{
			Base: responsex.NewBaseRespWithError(50400, "Token格式错误"),
		}, nil
	}

	claims, ok := atToken.Claims.(jwt.MapClaims)
	if !ok {
		return &authv1.LogoutResponse{
			Base: responsex.NewBaseRespWithError(50400, "Token格式错误"),
		}, nil
	}

	jti, _ := claims["jti"].(string)
	exp, _ := claims["exp"].(float64)

	// 2. 计算剩余 TTL，拉黑 AT（安全关键路径，写入失败必须返回错误）
	remainingTTL := int64(exp) - time.Now().Unix()
	if remainingTTL > 0 {
		blacklistKey := fmt.Sprintf("auth:at:blacklist:%s", jti)
		err := l.svcCtx.RedisClient.Set(l.ctx, blacklistKey, "1", time.Duration(remainingTTL)*time.Second).Err()
		if err != nil {
			l.Errorf("Logout: failed to blacklist AT jti=%s: %v", jti, err)
			return &authv1.LogoutResponse{
				Base: responsex.NewBaseRespWithError(50006, "注销失败，请稍后再试"),
			}, nil
		}
	}

	// 3. 清除当前设备 RT（AT 已拉黑，RT 删除失败不阻塞注销，仅记录日志）
	userID := in.UserId
	rtKey := fmt.Sprintf("auth:rt:%d:%s", userID, in.DeviceId)
	if err := l.svcCtx.RedisClient.Del(l.ctx, rtKey).Err(); err != nil {
		l.Errorf("Logout: failed to delete RT %s: %v", rtKey, err)
	}
	l.svcCtx.RedisClient.SRem(l.ctx, fmt.Sprintf("auth:rt:%d:devices", userID), in.DeviceId)

	// 4. 强踢全设备
	if in.KickAllDevices {
		devicesKey := fmt.Sprintf("auth:rt:%d:devices", userID)
		deviceIDs, _ := l.svcCtx.RedisClient.SMembers(l.ctx, devicesKey).Result()
		for _, devID := range deviceIDs {
			deviceRTKey := fmt.Sprintf("auth:rt:%d:%s", userID, devID)
			l.svcCtx.RedisClient.Del(l.ctx, deviceRTKey)
		}
		l.svcCtx.RedisClient.Del(l.ctx, devicesKey)
		l.Infof("Logout: kicked all devices for userId=%d, count=%d", userID, len(deviceIDs))
	}

	l.Infof("Logout: userId=%d, device=%s, jti=%s, remainingTTL=%ds", userID, in.DeviceId, jti, remainingTTL)

	return &authv1.LogoutResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
