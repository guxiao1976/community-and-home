package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 权限中心 API 服务配置
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	PermissionRpc zrpc.RpcClientConf
	DataSource    string // MySQL DSN（供 auto-discover 等管理功能使用）
	AesKey        string `json:",optional"` // AES 密钥（用于手机号解密）
}
