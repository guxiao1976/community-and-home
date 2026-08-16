package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 社区枢纽 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource      string             // MySQL DSN（数据库 community_hub_db）
	SysConfigRedis  redis.RedisConf    // 系统参数配置 Redis
	MasterDataRpc   zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
	ModerationRpc   zrpc.RpcClientConf // moderation-service gRPC 客户端配置
	ModerationRedis redis.RedisConf    // Redis for task queue
	PermissionRpc   zrpc.RpcClientConf // permission-service gRPC 客户端配置（AssertPublishScope / GetDataScopes / GetUserRoles）
	UserRpc         zrpc.RpcClientConf // user-service gRPC 客户端配置（GetUsersByIds 档案查询，R5）
	FileRpc         zrpc.RpcClientConf // file-service gRPC 客户端配置（GetFileUrl 附件绑定/重生，REQ-CPB-6）
	Kafka           KafkaConf          // Kafka 生产者配置（content-review topic）
}

// KafkaConf Kafka 生产者配置（Task 1.17）。
type KafkaConf struct {
	Brokers []string // broker 地址列表（docker-compose 网络内 kafka:9092）
	Topic   string   // content-review topic
	// RetryIntervalSeconds rescanner 定时重推周期（秒，默认 60）。
	RetryIntervalSeconds int
}
