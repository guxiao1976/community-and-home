package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	DataSource string
	Cache      cache.CacheConf
	Casbin     struct {
		ModelPath     string
		PolicyAdapter string
	}
	MinIO struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		BucketName      string
	}
}
