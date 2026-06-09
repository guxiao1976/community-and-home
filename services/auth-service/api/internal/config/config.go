package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 认证中心 REST API 配置
type Config struct {
	rest.RestConf
	AuthRpc        zrpc.RpcClientConf // Auth Service gRPC 客户端配置（etcd 发现）
	MasterDataRpc  zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
	JwtAuth        JwtAuthConfig      // JWT 配置（用于 rest.WithJwt）
	RedisAddr      string             // Redis 地址（用于短信验证码存储和限流）
	SysConfigRedis redis.RedisConf    // 系统参数配置 Redis（用于 sysconfig 客户端）
	RsaPublicKey      string          // RSA 公钥（PEM 格式，用于 /api/auth/public-key 端点）
	RsaPrivateKey     string          // RSA 私钥（PEM 格式，crypto.InitRSA 需要）— 已废弃，优先使用 RsaPrivateKeyPath
	RsaPrivateKeyPath string          // RSA 私钥文件路径（推荐，从文件读取 PEM）
}

// JwtAuthConfig JWT 配置（REST 层仅需 AccessSecret 用于验证签名）
type JwtAuthConfig struct {
	AccessSecret string // AT 签名密钥（与 RPC 端保持一致）
}
