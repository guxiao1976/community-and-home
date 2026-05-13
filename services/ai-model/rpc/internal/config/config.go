package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	PythonEngine PythonEngineConf
	Models       ModelsConf
	Cache        CacheConf `json:",optional"`
}

type PythonEngineConf struct {
	Host       string
	Port       int
	Timeout    int64 // 毫秒
	MaxRetries int
}

type ModelsConf struct {
	Text  ModelConf
	Image ModelConf
}

type ModelConf struct {
	Name    string
	Version string
	Enabled bool
}

type CacheConf struct {
	Enabled bool
	TTL     int64 // 秒
	Redis   struct {
		Host string
		Type string
	}
}
