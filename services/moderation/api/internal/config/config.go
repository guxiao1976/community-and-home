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
	DataSource    string
	LogDataSource string
	Cache         cache.CacheConf
	ACEngine      struct {
		SyncInterval      int
		PinyinExpand      bool
		PinyinMaxVariants int
		SplitDetect       bool
		SplitSeparators   string
	}
	Normalizer struct {
		IgnoreWidth        bool
		IgnoreCase         bool
		IgnoreChineseStyle bool
		IgnoreNumStyle     bool
		IgnoreEnglishStyle bool
		IgnoreRepeat       bool
	}
	SmallModel struct {
		Enable             bool
		Endpoint           string
		Model              string
		Timeout            int
		HighConfThreshold  float64
	}
	LargeModel struct {
		Enable  bool
		Endpoint string
		ApiKey  string
		Model   string
		Timeout int
	}
	ImageHash struct {
		Enable            bool
		HashSize          int
		SimilarThreshold  int
	}
}
