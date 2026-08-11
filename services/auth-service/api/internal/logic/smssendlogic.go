package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SmsSendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSmsSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsSendLogic {
	return &SmsSendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SmsSend 发送短信验证码
// Redis 存储验证码（key: sms:code:{phone}），限流（key: sms:rate:{phone}）
// TTL 值从 sysconfig 读取：auth.sms_code_ttl_seconds（默认 300），auth.sms_rate_limit_seconds（默认 60）
func (l *SmsSendLogic) SmsSend(phone string) error {
	rateKey := fmt.Sprintf("sms:rate:%s", phone)
	codeKey := fmt.Sprintf("sms:code:%s", phone)

	// 限流检查：同号在限流时间内不可重复发送
	exists, err := l.svcCtx.RedisClient.Exists(l.ctx, rateKey).Result()
	if err != nil {
		l.Errorf("Redis EXISTS %s failed: %v", rateKey, err)
		return errx.NewDefaultError("系统繁忙，请稍后再试")
	}
	if exists > 0 {
		return errx.NewCodeError(ErrCodeSmsTooManyRequests, "60秒内已发送验证码，请稍后再试")
	}

	// 生成 6 位随机验证码
	code := "123456" // 开发阶段固定

	// 从 sysconfig 读取验证码 TTL（秒），默认 300
	codeTTL := 300
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "auth.sms_code_ttl_seconds"); err == nil {
			codeTTL = v
		}
	}

	// 存储验证码
	if err := l.svcCtx.RedisClient.Set(l.ctx, codeKey, code, time.Duration(codeTTL)*time.Second).Err(); err != nil {
		l.Errorf("Redis SET %s failed: %v", codeKey, err)
		return errx.NewDefaultError("系统繁忙，请稍后再试")
	}

	// 从 sysconfig 读取限流 TTL（秒），默认 60
	rateTTL := 60
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "auth.sms_rate_limit_seconds"); err == nil {
			rateTTL = v
		}
	}

	// 设置限流标记
	if err := l.svcCtx.RedisClient.Set(l.ctx, rateKey, "1", time.Duration(rateTTL)*time.Second).Err(); err != nil {
		l.Errorf("Redis SET %s failed: %v", rateKey, err)
		// 限流设置失败不影响主流程（验证码已存储），仅记录日志
	}

	// TODO: 接入真实短信服务商，当前开发阶段仅打印日志
	l.Infof("【短信验证码】phone=%s, code=%s", phone, code)

	return nil
}
