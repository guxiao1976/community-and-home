package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 认证中心 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource     string          // MySQL DSN (数据库 auth)
	Cache          cache.CacheConf
	SysConfigRedis redis.RedisConf // 系统参数配置 Redis
	JwtAuth        JwtAuthConfig   // JWT 签发配置
	UserServiceRpc zrpc.RpcClientConf // User Service gRPC 客户端配置
	MasterDataRpc  zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
	RsaPrivateKey     string      // RSA 私钥（PEM 格式，用于解密手机号和密码）— 已废弃，优先使用 RsaPrivateKeyPath
	RsaPrivateKeyPath string      // RSA 私钥文件路径（推荐，从文件读取 PEM）
	RsaPublicKey      string      // RSA 公钥（PEM 格式，供 API Gateway 验签 AT）
	AesKey        string      // AES 密钥（用于加密手机号作为凭证标识符）
}

// JwtAuthConfig JWT 配置
type JwtAuthConfig struct {
	AccessSecret  string // AT 签名密钥
	AccessExpire  int64  // AT 过期时间（秒），默认 900（15 分钟）
	RefreshSecret string // RT 签名密钥
	RefreshExpire int64  // RT 过期时间（秒），默认 1296000（15 天）
}
